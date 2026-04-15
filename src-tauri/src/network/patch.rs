//! Surgical byte-level patching of network XML blobs.
//!
//! The parser stores link tag attributes as their raw attribute string (e.g.
//! `id="1" from="a" to="b" length="12.5"`). When the user edits a subset of
//! those fields, we rewrite only the matching attributes in place and leave
//! every other byte (including whitespace and attribute ordering) untouched.

use crate::ez::types::EzError;
use crate::utils::xml_escape;

use super::types::LinkTagEditInput;

/// Applies partial edits to a link tag attribute blob. Each `Some` field in
/// `edits` replaces the current value of that attribute byte-for-byte; other
/// attributes keep their original formatting. Attributes not present in the
/// blob are appended at the end of the string.
pub(crate) fn apply_link_tag_edits(
    original: &str,
    edits: &LinkTagEditInput,
) -> Result<String, EzError> {
    let updates: Vec<(&str, &str)> = [
        ("length", edits.length.as_deref()),
        ("freespeed", edits.freespeed.as_deref()),
        ("capacity", edits.capacity.as_deref()),
        ("permlanes", edits.permlanes.as_deref()),
        ("oneway", edits.oneway.as_deref()),
        ("modes", edits.modes.as_deref()),
    ]
    .into_iter()
    .filter_map(|(name, maybe)| maybe.map(|v| (name, v)))
    .collect();

    let mut out = String::with_capacity(original.len() + 32);
    let mut handled: Vec<&str> = Vec::with_capacity(updates.len());
    let bytes = original.as_bytes();
    let mut i = 0_usize;

    while i < bytes.len() {
        // Copy leading whitespace.
        let ws_start = i;
        while i < bytes.len() && bytes[i].is_ascii_whitespace() {
            i += 1;
        }
        out.push_str(&original[ws_start..i]);
        if i >= bytes.len() {
            break;
        }

        // Parse attribute name up to '='.
        let name_start = i;
        while i < bytes.len() && bytes[i] != b'=' && !bytes[i].is_ascii_whitespace() {
            i += 1;
        }
        let name = &original[name_start..i];
        if i >= bytes.len() || bytes[i] != b'=' {
            // Malformed; copy remainder verbatim.
            out.push_str(&original[name_start..]);
            break;
        }
        out.push_str(name);
        out.push('=');
        i += 1;

        if i >= bytes.len() || (bytes[i] != b'"' && bytes[i] != b'\'') {
            return Err(EzError::Io {
                message: format!("Malformed link tag blob near attribute '{name}'."),
            });
        }
        let quote = bytes[i];
        i += 1;
        let value_start = i;
        while i < bytes.len() && bytes[i] != quote {
            i += 1;
        }
        if i >= bytes.len() {
            return Err(EzError::Io {
                message: format!("Unterminated attribute '{name}' in link tag blob."),
            });
        }
        let original_value = &original[value_start..i];
        i += 1; // consume closing quote

        if let Some((_, new_value)) = updates.iter().find(|(n, _)| *n == name) {
            out.push(quote as char);
            out.push_str(&xml_escape(new_value));
            out.push(quote as char);
            handled.push(name);
        } else {
            out.push(quote as char);
            out.push_str(original_value);
            out.push(quote as char);
        }
    }

    // Append any updates that weren't present in the original.
    for (name, value) in &updates {
        if !handled.iter().any(|h| h == name) {
            out.push(' ');
            out.push_str(name);
            out.push_str("=\"");
            out.push_str(&xml_escape(value));
            out.push('"');
        }
    }

    Ok(out)
}

/// Replaces the first top-level `<attributes>` child inside `parent_xml` with
/// `new_blob`. If the parent has no existing `<attributes>` child, inserts
/// `new_blob` immediately before the parent's closing tag.
pub(crate) fn replace_attributes_child(
    parent_xml: &str,
    new_blob: Option<&str>,
) -> Result<String, EzError> {
    let (start, end) = find_top_level_attributes_span(parent_xml)?;
    match (start, new_blob) {
        (Some((lo, hi)), Some(blob)) => {
            let mut out = String::with_capacity(parent_xml.len() + blob.len());
            out.push_str(&parent_xml[..lo]);
            out.push_str(blob);
            out.push_str(&parent_xml[hi..]);
            Ok(out)
        }
        (Some((lo, hi)), None) => {
            let mut out = String::with_capacity(parent_xml.len());
            out.push_str(&parent_xml[..lo]);
            out.push_str(&parent_xml[hi..]);
            Ok(out)
        }
        (None, Some(blob)) => {
            // Insert before the closing tag. If the root is self-closing
            // (`<node .../>`), convert it into an open/close pair.
            if let Some(insert_at) = end {
                let mut out = String::with_capacity(parent_xml.len() + blob.len());
                out.push_str(&parent_xml[..insert_at]);
                out.push_str(blob);
                out.push_str(&parent_xml[insert_at..]);
                Ok(out)
            } else if let Some((open_end, name)) = find_self_closing_end(parent_xml) {
                let mut out = String::with_capacity(parent_xml.len() + blob.len() + name.len() + 4);
                out.push_str(&parent_xml[..open_end - 2]); // drop `/>`
                out.push('>');
                out.push_str(blob);
                out.push_str("</");
                out.push_str(name);
                out.push('>');
                Ok(out)
            } else {
                Err(EzError::Io {
                    message: "Unable to locate insertion point for <attributes>.".into(),
                })
            }
        }
        (None, None) => Ok(parent_xml.to_string()),
    }
}

/// Rewrites the `x` and `y` attributes of the root element in-place.
pub(crate) fn rewrite_node_xy(raw_xml: &str, x: f64, y: f64) -> Result<String, EzError> {
    let bytes = raw_xml.as_bytes();
    let Some(start) = bytes.iter().position(|&b| b == b'<') else {
        return Err(EzError::Io {
            message: "Node XML has no root element.".into(),
        });
    };
    let Some(end) = bytes[start..].iter().position(|&b| b == b'>').map(|p| p + start) else {
        return Err(EzError::Io {
            message: "Node XML root start tag is unterminated.".into(),
        });
    };

    let tag_range = &raw_xml[start..=end];
    let self_closing = tag_range.ends_with("/>");
    let inner = if self_closing {
        &tag_range[1..tag_range.len() - 2]
    } else {
        &tag_range[1..tag_range.len() - 1]
    };

    let x_str = format_decimal(x);
    let y_str = format_decimal(y);

    let updates: Vec<(&str, &str)> = vec![("x", &x_str), ("y", &y_str)];
    let mut name_end = 0_usize;
    let b = inner.as_bytes();
    while name_end < b.len() && !b[name_end].is_ascii_whitespace() {
        name_end += 1;
    }
    let name = &inner[..name_end];
    let attrs = &inner[name_end..];
    let patched_attrs = apply_tag_attribute_updates(attrs, &updates)?;

    let mut out = String::with_capacity(raw_xml.len() + 16);
    out.push_str(&raw_xml[..start]);
    out.push('<');
    out.push_str(name);
    out.push_str(&patched_attrs);
    if self_closing {
        out.push_str("/>");
    } else {
        out.push('>');
    }
    out.push_str(&raw_xml[end + 1..]);
    Ok(out)
}

fn apply_tag_attribute_updates(
    attrs: &str,
    updates: &[(&str, &str)],
) -> Result<String, EzError> {
    let mut out = String::with_capacity(attrs.len() + 16);
    let bytes = attrs.as_bytes();
    let mut i = 0_usize;
    let mut handled: Vec<&str> = Vec::new();

    while i < bytes.len() {
        let ws_start = i;
        while i < bytes.len() && bytes[i].is_ascii_whitespace() {
            i += 1;
        }
        out.push_str(&attrs[ws_start..i]);
        if i >= bytes.len() {
            break;
        }
        let name_start = i;
        while i < bytes.len() && bytes[i] != b'=' && !bytes[i].is_ascii_whitespace() {
            i += 1;
        }
        let name = &attrs[name_start..i];
        if i >= bytes.len() || bytes[i] != b'=' {
            out.push_str(&attrs[name_start..]);
            break;
        }
        out.push_str(name);
        out.push('=');
        i += 1;
        if i >= bytes.len() || (bytes[i] != b'"' && bytes[i] != b'\'') {
            return Err(EzError::Io {
                message: format!("Malformed attributes near '{name}'."),
            });
        }
        let quote = bytes[i];
        i += 1;
        let value_start = i;
        while i < bytes.len() && bytes[i] != quote {
            i += 1;
        }
        if i >= bytes.len() {
            return Err(EzError::Io {
                message: format!("Unterminated attribute '{name}'."),
            });
        }
        let original_value = &attrs[value_start..i];
        i += 1;
        if let Some((_, new_value)) = updates.iter().find(|(n, _)| *n == name) {
            out.push(quote as char);
            out.push_str(&xml_escape(new_value));
            out.push(quote as char);
            handled.push(name);
        } else {
            out.push(quote as char);
            out.push_str(original_value);
            out.push(quote as char);
        }
    }
    for (name, value) in updates {
        if !handled.iter().any(|h| h == name) {
            out.push(' ');
            out.push_str(name);
            out.push_str("=\"");
            out.push_str(&xml_escape(value));
            out.push('"');
        }
    }
    Ok(out)
}

fn find_top_level_attributes_span(parent_xml: &str) -> Result<(Option<(usize, usize)>, Option<usize>), EzError> {
    // Returns ((start, end) of <attributes>...</attributes> if present, insert position = start of parent's closing tag if non-self-closing).
    let bytes = parent_xml.as_bytes();
    // Skip root open tag.
    let open_start = bytes.iter().position(|&b| b == b'<').unwrap_or(0);
    let open_end = bytes[open_start..].iter().position(|&b| b == b'>').map(|p| p + open_start)
        .ok_or_else(|| EzError::Io { message: "Parent has no root element.".into() })?;
    if parent_xml[open_start..=open_end].ends_with("/>") {
        return Ok((None, None));
    }
    let mut i = open_end + 1;

    let close_marker = "</";
    let mut depth = 0_i32;
    while i < bytes.len() {
        if bytes[i] != b'<' {
            i += 1;
            continue;
        }
        if parent_xml[i..].starts_with(close_marker) && depth == 0 {
            return Ok((None, Some(i)));
        }
        // Find end of this tag
        let tag_end = bytes[i..].iter().position(|&b| b == b'>').map(|p| p + i)
            .ok_or_else(|| EzError::Io { message: "Unterminated child tag.".into() })?;
        let is_close = parent_xml[i..].starts_with("</");
        let is_self_closing = parent_xml[..=tag_end].ends_with("/>");
        // Capture element name
        let name_start = if is_close { i + 2 } else { i + 1 };
        let mut name_end = name_start;
        while name_end < bytes.len()
            && !bytes[name_end].is_ascii_whitespace()
            && bytes[name_end] != b'>'
            && bytes[name_end] != b'/'
        {
            name_end += 1;
        }
        let name = &parent_xml[name_start..name_end];

        if depth == 0 && !is_close && name == "attributes" {
            if is_self_closing {
                return Ok((Some((i, tag_end + 1)), None));
            }
            // Find matching close.
            let mut j = tag_end + 1;
            let mut inner_depth = 1_i32;
            while j < bytes.len() && inner_depth > 0 {
                if bytes[j] != b'<' { j += 1; continue; }
                let te = bytes[j..].iter().position(|&b| b == b'>').map(|p| p + j)
                    .ok_or_else(|| EzError::Io { message: "Unterminated tag inside <attributes>.".into() })?;
                let sc = parent_xml[..=te].ends_with("/>");
                let close = parent_xml[j..].starts_with("</");
                let ns = if close { j + 2 } else { j + 1 };
                let mut ne = ns;
                while ne < bytes.len()
                    && !bytes[ne].is_ascii_whitespace()
                    && bytes[ne] != b'>'
                    && bytes[ne] != b'/'
                {
                    ne += 1;
                }
                let nm = &parent_xml[ns..ne];
                if nm == "attributes" {
                    if close { inner_depth -= 1; }
                    else if !sc { inner_depth += 1; }
                }
                j = te + 1;
            }
            return Ok((Some((i, j)), None));
        }

        if is_close {
            depth -= 1;
        } else if !is_self_closing {
            depth += 1;
        }
        i = tag_end + 1;
    }
    Ok((None, None))
}

fn find_self_closing_end(parent_xml: &str) -> Option<(usize, &str)> {
    let bytes = parent_xml.as_bytes();
    let open_start = bytes.iter().position(|&b| b == b'<')?;
    let open_end = bytes[open_start..].iter().position(|&b| b == b'>').map(|p| p + open_start)?;
    if !parent_xml[open_start..=open_end].ends_with("/>") {
        return None;
    }
    let name_start = open_start + 1;
    let mut name_end = name_start;
    while name_end < bytes.len()
        && !bytes[name_end].is_ascii_whitespace()
        && bytes[name_end] != b'/'
        && bytes[name_end] != b'>'
    {
        name_end += 1;
    }
    Some((open_end + 1, &parent_xml[name_start..name_end]))
}

fn format_decimal(value: f64) -> String {
    format!("{value:.6}")
}
