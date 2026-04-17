use quick_xml::events::BytesStart;

use crate::ez::types::EzError;

pub(crate) fn required_attr_string(
    start: &BytesStart<'_>,
    key: &[u8],
    element_name: &str,
) -> Result<String, EzError> {
    optional_attr_string(start, key)?.ok_or_else(|| EzError::Io {
        message: format!(
            "Missing required attribute '{}' on <{element_name}>.",
            String::from_utf8_lossy(key)
        ),
    })
}

pub(crate) fn required_attr_f64(
    start: &BytesStart<'_>,
    key: &[u8],
    element_name: &str,
) -> Result<f64, EzError> {
    let value = required_attr_string(start, key, element_name)?;
    value.parse::<f64>().map_err(|err| EzError::Io {
        message: format!("Failed to parse numeric attribute on <{element_name}>: {err}"),
    })
}

pub(crate) fn optional_attr_string(
    start: &BytesStart<'_>,
    key: &[u8],
) -> Result<Option<String>, EzError> {
    let attr = start.try_get_attribute(key).map_err(|err| EzError::Io {
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

pub(crate) fn optional_attr_bool(
    start: &BytesStart<'_>,
    key: &[u8],
) -> Result<Option<bool>, EzError> {
    let value = optional_attr_string(start, key)?;
    match value.as_deref() {
        Some("true") => Ok(Some(true)),
        Some("false") => Ok(Some(false)),
        Some(other) => Err(EzError::Io {
            message: format!(
                "Invalid boolean value '{}' for attribute '{}'",
                other,
                String::from_utf8_lossy(key)
            ),
        }),
        None => Ok(None),
    }
}
