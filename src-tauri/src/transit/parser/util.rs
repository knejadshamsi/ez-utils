use std::io::BufRead;

use quick_xml::{events::Event, reader::Reader};

use crate::ez::types::EzError;

pub(super) fn consume_to_end<R: BufRead>(
    reader: &mut Reader<R>,
    event_buf: &mut Vec<u8>,
    name: &[u8],
) -> Result<(), EzError> {
    loop {
        event_buf.clear();
        let event = reader
            .read_event_into(event_buf)
            .map(|event| event.into_owned())
            .map_err(|err| EzError::Io {
                message: format!("Failed to parse transit XML: {err}"),
            })?;

        match event {
            Event::End(ref end) if end.local_name().as_ref() == name => return Ok(()),
            Event::Eof => {
                return Err(EzError::Io {
                    message: format!(
                        "Unexpected EOF before </{}>.",
                        String::from_utf8_lossy(name)
                    ),
                });
            }
            _ => {}
        }
    }
}

pub(super) fn read_text_content<R: BufRead>(
    reader: &mut Reader<R>,
    event_buf: &mut Vec<u8>,
) -> Result<String, EzError> {
    let mut text = String::new();
    loop {
        event_buf.clear();
        let event = reader
            .read_event_into(event_buf)
            .map(|event| event.into_owned())
            .map_err(|err| EzError::Io {
                message: format!("Failed to parse transit XML: {err}"),
            })?;

        match event {
            Event::Text(t) => {
                text.push_str(
                    std::str::from_utf8(t.as_ref()).map_err(|err| EzError::Io {
                        message: format!("Text content was not valid UTF-8: {err}"),
                    })?,
                );
            }
            Event::End(_) => break,
            Event::Eof => {
                return Err(EzError::Io {
                    message: "Unexpected EOF while reading text content.".into(),
                });
            }
            _ => {}
        }
    }
    Ok(text.trim().to_string())
}

pub(super) fn sqlite_err(err: rusqlite::Error) -> EzError {
    EzError::Io {
        message: format!("Failed to write transit SQLite rows: {err}"),
    }
}
