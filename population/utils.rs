use anyhow::{Context, Result};
use geo::Point;
use proj::Proj;
use std::env;
use std::path::Path;

pub fn transform_coordinates(x: f64, y: f64) -> Result<(f64, f64)> {
    let proj = Proj::new_known_crs(
        "EPSG:2950",
        "EPSG:4326",
        None
    ).context("Failed to create coordinate transformer")?;

    proj.convert((x, y))
        .context("Failed to transform coordinates")
}

pub fn create_point_from_coordinates(x: f64, y: f64) -> Result<Point<f64>> {
    let (lon, lat) = transform_coordinates(x, y)?;
    Ok(Point::new(lon, lat))
}

pub fn parse_time(time_str: &str) -> Option<String> {
    if time_str.is_empty() {
        return None;
    }

    let parts: Vec<&str> = time_str.trim().split(':').collect();
    if parts.len() != 3 {
        return None;
    }

    let hours: i32 = parts[0].parse().ok()?;
    let minutes: i32 = parts[1].parse().ok()?;
    let seconds: i32 = parts[2].parse().ok()?;

    if minutes > 59 || seconds > 59 {
        return None;
    }

    Some(format!("{:02}:{:02}:{:02}", hours, minutes, seconds))
}

pub fn get_total_population() -> Result<i32> {
    env::var("TOTAL_POPULATION")
        .context("TOTAL_POPULATION not set in environment")?
        .parse()
        .context("Failed to parse TOTAL_POPULATION as integer")
}

pub fn validate_output_dir(output_dir: &Path) -> Result<()> {
    if !output_dir.exists() {
        std::fs::create_dir_all(output_dir)
            .context("Failed to create output directory")?;
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_time() {
        assert_eq!(parse_time("12:34:56"), Some("12:34:56".to_string()));
        assert_eq!(parse_time("1:2:3"), Some("01:02:03".to_string()));
        assert_eq!(parse_time(""), None);
        assert_eq!(parse_time("12:60:00"), None);
        assert_eq!(parse_time("12:00:60"), None);
        assert_eq!(parse_time("invalid"), None);
    }
}
