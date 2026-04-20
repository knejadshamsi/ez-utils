use quick_xml::{
    events::{BytesStart, Event},
    reader::Reader,
};

use crate::ez::types::EzError;
use crate::utils::xml_escape;

use crate::projection::Projector;

use super::{
    capture::{capture_empty_element, capture_started_element},
    types::{ApplyPlanEditsInput, PlanActivityEditInput, PlanLegEditInput},
};

pub(crate) fn apply_plan_edits_to_blob(
    original_blob: &str,
    edits: &ApplyPlanEditsInput,
    projector: &Projector,
) -> Result<String, EzError> {
    let parsed = parse_plan_segments(original_blob)?;
    let mut out = String::with_capacity(original_blob.len() + 256);

    out.push_str(&parsed.root_start);
    out.push_str(&parsed.prefix);

    for (index, activity) in edits.activities.iter().enumerate() {
        out.push_str(&render_activity(activity, &parsed, projector)?);
        if let Some(leg) = edits.legs.get(index) {
            out.push_str(&render_leg(leg, &parsed)?);
        }
    }

    out.push_str(&parsed.suffix);
    out.push_str(&parsed.root_end);
    Ok(out)
}

fn render_activity(
    activity: &PlanActivityEditInput,
    parsed: &ParsedPlanSegments,
    projector: &Projector,
) -> Result<String, EzError> {
    if !activity.dirty {
        if let Some(original_index) = activity.original_index {
            if let Some(raw) = parsed.activities.get(original_index as usize) {
                return Ok(raw.clone());
            }
        }
    }

    let mut attrs = vec![("type", xml_escape(&activity.type_name))];
    if let Some(start_time) = activity.start_time.as_deref().filter(|value| !value.is_empty()) {
        attrs.push(("start_time", xml_escape(start_time)));
    }
    if let Some(end_time) = activity.end_time.as_deref().filter(|value| !value.is_empty()) {
        attrs.push(("end_time", xml_escape(end_time)));
    }
    if let (Some(lng), Some(lat)) = (activity.lng, activity.lat) {
        let native = projector.unproject_lng_lat(lng, lat)?;
        attrs.push(("x", format_decimal(native.x)));
        attrs.push(("y", format_decimal(native.y)));
    }

    Ok(render_empty_element("activity", &attrs))
}

fn render_leg(
    leg: &PlanLegEditInput,
    parsed: &ParsedPlanSegments,
) -> Result<String, EzError> {
    if !leg.dirty {
        if let Some(original_index) = leg.original_index {
            if let Some(raw) = parsed.legs.get(original_index as usize) {
                return Ok(raw.clone());
            }
        }
    }

    let mut attrs = vec![("mode", xml_escape(&leg.mode))];
    if let Some(travel_time) = leg.travel_time.as_deref().filter(|value| !value.is_empty()) {
        attrs.push(("trav_time", xml_escape(travel_time)));
    }

    Ok(render_empty_element("leg", &attrs))
}

fn render_empty_element(name: &str, attrs: &[(&str, String)]) -> String {
    let mut out = String::new();
    out.push('<');
    out.push_str(name);
    for (attr_name, value) in attrs {
        out.push(' ');
        out.push_str(attr_name);
        out.push_str("=\"");
        out.push_str(value);
        out.push('"');
    }
    out.push_str("/>");
    out
}

fn format_decimal(value: f64) -> String {
    format!("{value:.6}")
}

#[derive(Debug)]
struct ParsedPlanSegments {
    root_start: String,
    prefix: String,
    suffix: String,
    root_end: String,
    activities: Vec<String>,
    legs: Vec<String>,
}

fn parse_plan_segments(plan_blob: &str) -> Result<ParsedPlanSegments, EzError> {
    let mut reader = Reader::from_reader(plan_blob.as_bytes());
    reader.config_mut().trim_text(false);

    let mut event_buf = Vec::new();
    let mut capture_buf = Vec::new();

    let mut root_start = String::new();
    let root_end: String;
    let mut prefix = String::new();
    let mut suffix = String::new();
    let mut activities = Vec::new();
    let mut legs = Vec::new();
    let mut seen_dynamic = false;

    loop {
        event_buf.clear();
        let event = reader
            .read_event_into(&mut event_buf)
            .map(|event| event.into_owned())
            .map_err(|err| EzError::Io {
                message: format!("Failed to parse stored plan XML: {err}"),
            })?;

        match event {
            Event::Start(ref start) if start.local_name().into_inner() == b"plan" => {
                root_start = capture_root_start(start)?;
            }
            Event::End(ref end) if end.local_name().as_ref() == b"plan" => {
                root_end = format!("</{}>", String::from_utf8_lossy(end.as_ref()));
                break;
            }
            Event::Start(ref start) if is_dynamic_name(start) => {
                let raw = capture_started_element(
                    &mut reader,
                    start,
                    &mut event_buf,
                    &mut capture_buf,
                )?;
                if start.local_name().into_inner() == b"activity" {
                    activities.push(raw);
                } else {
                    legs.push(raw);
                }
                seen_dynamic = true;
            }
            Event::Empty(ref start) if is_dynamic_name(start) => {
                let raw = capture_empty_element(start)?;
                if start.local_name().into_inner() == b"activity" {
                    activities.push(raw);
                } else {
                    legs.push(raw);
                }
                seen_dynamic = true;
            }
            Event::Text(ref text) => append_static_segment(&mut prefix, &mut suffix, seen_dynamic, text.as_ref()),
            Event::Comment(ref comment) => append_wrapped_static(
                &mut prefix,
                &mut suffix,
                seen_dynamic,
                "<!--",
                comment.as_ref(),
                "-->",
            ),
            Event::CData(ref cdata) => append_wrapped_static(
                &mut prefix,
                &mut suffix,
                seen_dynamic,
                "<![CDATA[",
                cdata.as_ref(),
                "]]>",
            ),
            Event::PI(ref pi) => append_wrapped_static(
                &mut prefix,
                &mut suffix,
                seen_dynamic,
                "<?",
                pi.as_ref(),
                "?>",
            ),
            Event::GeneralRef(ref reference) => append_wrapped_static(
                &mut prefix,
                &mut suffix,
                seen_dynamic,
                "&",
                reference.as_ref(),
                ";",
            ),
            Event::Empty(ref start) => {
                let raw = capture_empty_element(start)?;
                append_static_segment(&mut prefix, &mut suffix, seen_dynamic, raw.as_bytes());
            }
            Event::Start(ref start) => {
                let raw = capture_started_element(
                    &mut reader,
                    start,
                    &mut event_buf,
                    &mut capture_buf,
                )?;
                append_static_segment(&mut prefix, &mut suffix, seen_dynamic, raw.as_bytes());
            }
            Event::Decl(ref decl) => append_wrapped_static(
                &mut prefix,
                &mut suffix,
                seen_dynamic,
                "<?",
                decl.as_ref(),
                "?>",
            ),
            Event::DocType(ref doc) => append_wrapped_static(
                &mut prefix,
                &mut suffix,
                seen_dynamic,
                "<!DOCTYPE ",
                doc.as_ref(),
                ">",
            ),
            Event::Eof => {
                return Err(EzError::Io {
                    message: "Unexpected EOF while patching plan XML.".into(),
                });
            }
            _ => {}
        }
    }

    Ok(ParsedPlanSegments {
        root_start,
        prefix,
        suffix,
        root_end,
        activities,
        legs,
    })
}

fn capture_root_start(start: &BytesStart<'_>) -> Result<String, EzError> {
    let mut out = Vec::with_capacity(start.len() + 2);
    out.push(b'<');
    out.extend_from_slice(start.as_ref());
    out.push(b'>');
    String::from_utf8(out).map_err(|err| EzError::Io {
        message: format!("Plan root start tag was not valid UTF-8: {err}"),
    })
}

fn is_dynamic_name(start: &BytesStart<'_>) -> bool {
    matches!(start.local_name().into_inner(), b"activity" | b"leg")
}

fn append_static_segment(prefix: &mut String, suffix: &mut String, seen_dynamic: bool, bytes: &[u8]) {
    let target = if seen_dynamic { suffix } else { prefix };
    target.push_str(&String::from_utf8_lossy(bytes));
}

fn append_wrapped_static(
    prefix: &mut String,
    suffix: &mut String,
    seen_dynamic: bool,
    before: &str,
    bytes: &[u8],
    after: &str,
) {
    let target = if seen_dynamic { suffix } else { prefix };
    target.push_str(before);
    target.push_str(&String::from_utf8_lossy(bytes));
    target.push_str(after);
}
