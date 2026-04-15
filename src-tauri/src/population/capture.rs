use std::io::BufRead;

use quick_xml::{
    events::{BytesStart, Event},
    name::QName,
    reader::Reader,
};

use crate::ez::types::EzError;

pub(crate) fn capture_empty_element(start: &BytesStart<'_>) -> Result<String, EzError> {
    let mut out: Vec<u8> = Vec::with_capacity(start.len() + 3);
    out.push(b'<');
    out.extend_from_slice(start.as_ref());
    out.extend_from_slice(b"/>");
    String::from_utf8(out).map_err(|err| EzError::Io {
        message: format!("Captured population XML was not valid UTF-8: {err}"),
    })
}

pub(crate) fn capture_started_element<R: BufRead>(
    reader: &mut Reader<R>,
    start: &BytesStart<'_>,
    event_buf: &mut Vec<u8>,
    out: &mut Vec<u8>,
) -> Result<String, EzError> {
    out.clear();
    append_start(start, out);

    let end_name = start.name();
    let mut depth = 0_i32;

    loop {
        event_buf.clear();
        let event = reader
            .read_event_into(event_buf)
            .map_err(|err| EzError::Io {
                message: format!("Failed to parse population XML: {err}"),
            })?;

        match event {
            Event::Start(ref inner) => {
                append_start(inner, out);
                if inner.name() == end_name {
                    depth += 1;
                }
            }
            Event::End(ref inner) => {
                append_end(inner.name(), inner.as_ref(), out);
                if inner.name() == end_name {
                    if depth == 0 {
                        break;
                    }
                    depth -= 1;
                }
            }
            Event::Eof => {
                return Err(EzError::Io {
                    message: "Unexpected EOF while capturing population XML subtree.".into(),
                });
            }
            other => append_event_bytes(&other, out),
        }
    }

    String::from_utf8(out.clone()).map_err(|err| EzError::Io {
        message: format!("Captured population XML was not valid UTF-8: {err}"),
    })
}

/// Serializes a single XML event verbatim into `out`. Covers every event
/// variant quick_xml can emit except structural Start/End pairs (callers
/// that need to track depth must handle those themselves; see
/// `capture_started_element`). `Event::Eof` is a no-op.
pub(crate) fn append_event_bytes(event: &Event<'_>, out: &mut Vec<u8>) {
    match event {
        Event::Start(start) => append_start(start, out),
        Event::Empty(start) => append_empty(start, out),
        Event::End(end) => append_end(end.name(), end.as_ref(), out),
        Event::Text(inner) => out.extend_from_slice(inner.as_ref()),
        Event::CData(inner) => {
            out.extend_from_slice(b"<![CDATA[");
            out.extend_from_slice(inner.as_ref());
            out.extend_from_slice(b"]]>");
        }
        Event::Comment(inner) => {
            out.extend_from_slice(b"<!--");
            out.extend_from_slice(inner.as_ref());
            out.extend_from_slice(b"-->");
        }
        Event::Decl(inner) => {
            out.extend_from_slice(b"<?");
            out.extend_from_slice(inner.as_ref());
            out.extend_from_slice(b"?>");
        }
        Event::PI(inner) => {
            out.extend_from_slice(b"<?");
            out.extend_from_slice(inner.as_ref());
            out.extend_from_slice(b"?>");
        }
        Event::DocType(inner) => {
            out.extend_from_slice(b"<!DOCTYPE ");
            out.extend_from_slice(inner.as_ref());
            out.push(b'>');
        }
        Event::GeneralRef(inner) => {
            out.push(b'&');
            out.extend_from_slice(inner.as_ref());
            out.push(b';');
        }
        Event::Eof => {}
    }
}

fn append_start(start: &BytesStart<'_>, out: &mut Vec<u8>) {
    out.push(b'<');
    out.extend_from_slice(start.as_ref());
    out.push(b'>');
}

fn append_empty(start: &BytesStart<'_>, out: &mut Vec<u8>) {
    out.push(b'<');
    out.extend_from_slice(start.as_ref());
    out.extend_from_slice(b"/>");
}

fn append_end(name: QName<'_>, raw: &[u8], out: &mut Vec<u8>) {
    let _ = name;
    out.extend_from_slice(b"</");
    out.extend_from_slice(raw);
    out.push(b'>');
}
