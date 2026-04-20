//! Population import parser.
//!
//! `plan_id` is system-generated as `format!("p{plan_index}")`, where `plan_index`
//! is the zero-based source order of `<plan>` elements inside a single `<person>`.

use std::{
    fs::{self, File},
    io::BufReader,
    path::Path,
};

use quick_xml::{
    events::{BytesStart, Event},
    reader::Reader,
};
use rusqlite::{params, Connection};

use crate::ez::{
    fs::{db_file_path, postamble_file_path, preamble_file_path},
    types::EzError,
};

use crate::projection::{CrsConfig, Projector};

use super::{
    capture::{append_event_bytes, capture_empty_element, capture_started_element},
    types::{
        ActivityCoordRecord, PersonAttributesRecord, PersonState, PlanRecord, RawActivityCoord,
    },
};

pub(crate) fn run(
    work_dir: &Path,
    file_path: &Path,
    source_name: &str,
    crs_config: &CrsConfig,
) -> Result<(), EzError> {
    let file = File::open(file_path).map_err(|err| EzError::Io {
        message: format!("Failed to open population XML: {err}"),
    })?;
    let mut reader = Reader::from_reader(BufReader::new(file));
    reader.config_mut().trim_text(false);
    reader.config_mut().trim_markup_names_in_closing_tags = false;

    let projector = Projector::new(crs_config)?;
    let mut event_buf = Vec::new();
    let mut capture_buf = Vec::new();
    let mut preamble_buf: Vec<u8> = Vec::new();
    let mut postamble_buf: Vec<u8> = Vec::new();
    let mut in_root = false;
    let mut root_closed = false;

    let mut connection =
        Connection::open(db_file_path(work_dir, source_name)).map_err(|err| EzError::Io {
            message: format!("Failed to open population SQLite database: {err}"),
        })?;
    let tx = connection.transaction().map_err(|err| EzError::Io {
        message: format!("Failed to begin population import transaction: {err}"),
    })?;

    {
        let mut insert_attributes = tx
            .prepare(
                "INSERT INTO person_attributes (person_id, attributes_blob) VALUES (?1, ?2)",
            )
            .map_err(sqlite_err)?;
        let mut insert_plan = tx
            .prepare(
                "INSERT INTO person_plans (person_id, plan_id, plan_index, selected, plan_blob)
                 VALUES (?1, ?2, ?3, ?4, ?5)",
            )
            .map_err(sqlite_err)?;
        let mut insert_coord = tx
            .prepare(
                "INSERT INTO plan_activity_coords
                 (person_id, plan_id, plan_index, activity_index, lng, lat)
                 VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
            )
            .map_err(sqlite_err)?;

        let mut current_person: Option<PersonState> = None;

        loop {
            event_buf.clear();
            let event = reader
                .read_event_into(&mut event_buf)
                .map(|event| event.into_owned())
                .map_err(|err| EzError::Io {
                    message: format!("Failed to parse population XML: {err}"),
                })?;

            // Phase 1: capture everything before the root <population> tag
            // verbatim into the preamble buffer. Includes the root Start tag
            // itself so the preamble file ends with `<population ...>` ready
            // to concatenate.
            if !in_root {
                match &event {
                    Event::Start(start)
                        if start.local_name().into_inner() == b"population" =>
                    {
                        append_event_bytes(&event, &mut preamble_buf);
                        in_root = true;
                        continue;
                    }
                    Event::Eof => {
                        return Err(EzError::Io {
                            message: "Population XML ended before root <population> element."
                                .into(),
                        });
                    }
                    _ => {
                        append_event_bytes(&event, &mut preamble_buf);
                        continue;
                    }
                }
            }

            // Phase 3: capture everything after the root </population> tag
            // verbatim into the postamble buffer.
            if root_closed {
                match &event {
                    Event::Eof => break,
                    _ => {
                        append_event_bytes(&event, &mut postamble_buf);
                        continue;
                    }
                }
            }

            // Phase 2: body parsing.
            match event {
                Event::Start(ref start) if start.local_name().into_inner() == b"person" => {
                    current_person = Some(PersonState {
                        person_id: required_attr_string(start, b"id", "person")?,
                        next_plan_index: 0,
                        wrote_attributes: false,
                    });
                }
                Event::Empty(ref start) if start.local_name().into_inner() == b"person" => {
                    let person_id = required_attr_string(start, b"id", "person")?;
                    insert_person_attributes(
                        &mut insert_attributes,
                        PersonAttributesRecord {
                            person_id,
                            attributes_blob: None,
                        },
                    )?;
                }
                Event::Start(ref start) if start.local_name().into_inner() == b"attributes" => {
                    let person = current_person.as_mut().ok_or_else(|| EzError::Io {
                        message: "Encountered <attributes> outside of <person>.".into(),
                    })?;
                    let blob = capture_started_element(
                        &mut reader,
                        start,
                        &mut event_buf,
                        &mut capture_buf,
                    )?;
                    insert_person_attributes(
                        &mut insert_attributes,
                        PersonAttributesRecord {
                            person_id: person.person_id.clone(),
                            attributes_blob: Some(blob),
                        },
                    )?;
                    person.wrote_attributes = true;
                }
                Event::Empty(ref start) if start.local_name().into_inner() == b"attributes" => {
                    let person = current_person.as_mut().ok_or_else(|| EzError::Io {
                        message: "Encountered <attributes/> outside of <person>.".into(),
                    })?;
                    let blob = capture_empty_element(start)?;
                    insert_person_attributes(
                        &mut insert_attributes,
                        PersonAttributesRecord {
                            person_id: person.person_id.clone(),
                            attributes_blob: Some(blob),
                        },
                    )?;
                    person.wrote_attributes = true;
                }
                Event::Start(ref start) if start.local_name().into_inner() == b"plan" => {
                    let person = current_person.as_mut().ok_or_else(|| EzError::Io {
                        message: "Encountered <plan> outside of <person>.".into(),
                    })?;
                    let plan_index = person.next_plan_index;
                    let plan_id = format!("p{plan_index}");
                    let selected = selected_value(start)?;
                    let plan_blob = capture_started_element(
                        &mut reader,
                        start,
                        &mut event_buf,
                        &mut capture_buf,
                    )?;
                    let plan_record = PlanRecord {
                        person_id: person.person_id.clone(),
                        plan_id: plan_id.clone(),
                        plan_index,
                        selected,
                        plan_blob,
                    };
                    insert_plan_record(&mut insert_plan, &plan_record)?;
                    insert_plan_activity_coords(
                        &mut insert_coord,
                        &projector,
                        &plan_record.person_id,
                        &plan_record.plan_id,
                        plan_record.plan_index,
                        &plan_record.plan_blob,
                    )?;
                    person.next_plan_index += 1;
                }
                Event::Empty(ref start) if start.local_name().into_inner() == b"plan" => {
                    let person = current_person.as_mut().ok_or_else(|| EzError::Io {
                        message: "Encountered <plan/> outside of <person>.".into(),
                    })?;
                    let plan_index = person.next_plan_index;
                    let plan_record = PlanRecord {
                        person_id: person.person_id.clone(),
                        plan_id: format!("p{plan_index}"),
                        plan_index,
                        selected: selected_value(start)?,
                        plan_blob: capture_empty_element(start)?,
                    };
                    insert_plan_record(&mut insert_plan, &plan_record)?;
                    person.next_plan_index += 1;
                }
                Event::End(ref end) if end.local_name().as_ref() == b"person" => {
                    let person = current_person.take().ok_or_else(|| EzError::Io {
                        message: "Encountered </person> without an open <person>.".into(),
                    })?;
                    if !person.wrote_attributes {
                        insert_person_attributes(
                            &mut insert_attributes,
                            PersonAttributesRecord {
                                person_id: person.person_id,
                                attributes_blob: None,
                            },
                        )?;
                    }
                }
                Event::End(ref end) if end.local_name().as_ref() == b"population" => {
                    postamble_buf.push(b'<');
                    postamble_buf.push(b'/');
                    postamble_buf.extend_from_slice(end.as_ref());
                    postamble_buf.push(b'>');
                    root_closed = true;
                }
                Event::Eof => {
                    return Err(EzError::Io {
                        message: "Population XML ended before </population> closing tag.".into(),
                    });
                }
                _ => {}
            }
        }
    }

    tx.commit().map_err(|err| EzError::Io {
        message: format!("Failed to commit population import transaction: {err}"),
    })?;

    fs::write(preamble_file_path(work_dir, source_name), &preamble_buf).map_err(|err| {
        EzError::Io {
            message: format!("Failed to write population preamble file: {err}"),
        }
    })?;
    fs::write(postamble_file_path(work_dir, source_name), &postamble_buf).map_err(|err| {
        EzError::Io {
            message: format!("Failed to write population postamble file: {err}"),
        }
    })?;

    Ok(())
}

fn insert_person_attributes(
    stmt: &mut rusqlite::Statement<'_>,
    record: PersonAttributesRecord,
) -> Result<(), EzError> {
    stmt.execute(params![record.person_id, record.attributes_blob])
        .map_err(sqlite_err)?;
    Ok(())
}

fn insert_plan_record(
    stmt: &mut rusqlite::Statement<'_>,
    record: &PlanRecord,
) -> Result<(), EzError> {
    stmt.execute(params![
        record.person_id,
        record.plan_id,
        record.plan_index,
        record.selected,
        record.plan_blob
    ])
    .map_err(sqlite_err)?;
    Ok(())
}

fn insert_plan_activity_coords(
    stmt: &mut rusqlite::Statement<'_>,
    projector: &Projector,
    person_id: &str,
    plan_id: &str,
    plan_index: i64,
    plan_blob: &str,
) -> Result<(), EzError> {
    for record in parse_plan_activity_coords(person_id, plan_id, plan_index, plan_blob, projector)? {
        stmt.execute(params![
            record.person_id,
            record.plan_id,
            record.plan_index,
            record.activity_index,
            record.lng,
            record.lat
        ])
        .map_err(sqlite_err)?;
    }
    Ok(())
}

fn parse_plan_activity_coords(
    person_id: &str,
    plan_id: &str,
    plan_index: i64,
    plan_blob: &str,
    projector: &Projector,
) -> Result<Vec<ActivityCoordRecord>, EzError> {
    let mut reader = Reader::from_reader(plan_blob.as_bytes());
    reader.config_mut().trim_text(false);

    let mut buf = Vec::new();
    let mut activity_index = 0_i64;
    let mut rows = Vec::new();

    loop {
        buf.clear();
        let event = reader
            .read_event_into(&mut buf)
            .map_err(|err| EzError::Io {
                message: format!("Failed to parse captured plan XML: {err}"),
            })?;

        match event {
            Event::Start(ref start) | Event::Empty(ref start)
                if start.local_name().into_inner() == b"activity" =>
            {
                let raw = parse_activity_xy(start, activity_index)?;
                activity_index += 1;

                if let Some(raw) = raw {
                    let projected = projector.project_xy(raw.x, raw.y)?;
                    rows.push(ActivityCoordRecord {
                        person_id: person_id.to_string(),
                        plan_id: plan_id.to_string(),
                        plan_index,
                        activity_index: raw.activity_index,
                        lng: projected.lng,
                        lat: projected.lat,
                    });
                }
            }
            Event::Eof => break,
            _ => {}
        }
    }

    Ok(rows)
}

fn parse_activity_xy(
    start: &BytesStart<'_>,
    activity_index: i64,
) -> Result<Option<RawActivityCoord>, EzError> {
    let x = optional_attr_f64(start, b"x")?;
    let y = optional_attr_f64(start, b"y")?;

    match (x, y) {
        (Some(x), Some(y)) => Ok(Some(RawActivityCoord {
            activity_index,
            x,
            y,
        })),
        (None, None) => Ok(None),
        _ => Err(EzError::Io {
            message: "Activity had only one of x/y; both are required to project coordinates."
                .into(),
        }),
    }
}

fn selected_value(start: &BytesStart<'_>) -> Result<i64, EzError> {
    let selected = optional_attr_string(start, b"selected")?;
    Ok(match selected.as_deref() {
        Some("yes") => 1,
        _ => 0,
    })
}

fn required_attr_string(
    start: &BytesStart<'_>,
    key: &[u8],
    element_name: &str,
) -> Result<String, EzError> {
    optional_attr_string(start, key)?.ok_or_else(|| EzError::Io {
        message: format!("Missing required attribute on <{element_name}>."),
    })
}

fn optional_attr_string(
    start: &BytesStart<'_>,
    key: &[u8],
) -> Result<Option<String>, EzError> {
    let attr = start
        .try_get_attribute(key)
        .map_err(|err| EzError::Io {
            message: format!("Failed to read XML attribute: {err}"),
        })?;

    attr.map(|attr| {
        std::str::from_utf8(attr.value.as_ref())
            .map(|value| value.to_string())
            .map_err(|err| EzError::Io {
                message: format!("XML attribute was not valid UTF-8: {err}"),
            })
    })
    .transpose()
}

fn optional_attr_f64(start: &BytesStart<'_>, key: &[u8]) -> Result<Option<f64>, EzError> {
    optional_attr_string(start, key)?
        .map(|value| {
            value.parse::<f64>().map_err(|err| EzError::Io {
                message: format!("Failed to parse numeric XML attribute '{value}': {err}"),
            })
        })
        .transpose()
}

fn sqlite_err(err: rusqlite::Error) -> EzError {
    EzError::Io {
        message: format!("Failed to write population SQLite rows: {err}"),
    }
}
