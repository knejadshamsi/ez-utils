//! Small shared utilities used across the crate.

/// Escapes the five XML-significant characters so `value` is safe to embed
/// inside an XML attribute value or element text. Replaces `&`, `"`, `<`, `>`
/// with their entity references (apostrophe is not escaped because double
/// quotes are used as the attribute delimiter throughout this codebase).
pub fn xml_escape(value: &str) -> String {
    value
        .replace('&', "&amp;")
        .replace('"', "&quot;")
        .replace('<', "&lt;")
        .replace('>', "&gt;")
}
