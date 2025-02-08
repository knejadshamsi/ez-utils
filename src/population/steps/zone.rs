use anyhow::Result;
use std::fs;
use geo::{MultiPoint, Point, Polygon, BoundingRect};
use geo::algorithm::convex_hull::ConvexHull;
use serde::Serialize;
use crate::population::core::models::{ProcessingPaths, Person};

#[derive(Serialize)]
struct AreaPolygon {
    id: usize,
    coordinates: Vec<[f64; 2]>,
}

#[derive(Serialize)]
struct BinPolygons {
    bins: Vec<AreaPolygon>,
}

fn create_boundary_polygon(persons: &[Person]) -> Polygon<f64> {
    let points: Vec<Point<f64>> = persons.iter()
        .filter_map(|p| p.activities.first())
        .filter_map(|a| a.coordinates)
        .collect();

    MultiPoint::new(points).convex_hull()
}

fn create_area_bins(boundary: &Polygon<f64>, bin_count: usize) -> Vec<Polygon<f64>> {
    let bbox = boundary.bounding_rect().unwrap();
    let width = bbox.width();
    let height = bbox.height();
    
    let cols = (bin_count as f64).sqrt().ceil() as usize;
    let rows = (bin_count + cols - 1) / cols;
    
    let cell_width = width / cols as f64;
    let cell_height = height / rows as f64;
    
    let mut bins = Vec::new();
    
    for row in 0..rows {
        for col in 0..cols {
            if bins.len() >= bin_count {
                break;
            }
            
            let min_x = bbox.min().x + col as f64 * cell_width;
            let min_y = bbox.min().y + row as f64 * cell_height;
            
            let points = vec![
                Point::new(min_x, min_y),
                Point::new(min_x + cell_width, min_y),
                Point::new(min_x + cell_width, min_y + cell_height),
                Point::new(min_x, min_y + cell_height),
                Point::new(min_x, min_y),
            ];
            
            bins.push(Polygon::new(points.into(), vec![]));
        }
    }
    
    bins
}

fn polygon_to_serializable(polygon: &Polygon<f64>, id: usize) -> AreaPolygon {
    AreaPolygon {
        id,
        coordinates: polygon.exterior()
            .points()
            .map(|p| [p.x(), p.y()])
            .collect(),
    }
}

pub fn process_step_3(persons: Vec<Person>, paths: &ProcessingPaths) -> Result<Vec<Person>> {
    // Create full area polygon
    let boundary = create_boundary_polygon(&persons);
    let full_area = polygon_to_serializable(&boundary, 0); // Use 0 as id for full area
    
    // Create bins (using 16 as example, adjust as needed)
    let bin_polygons = create_area_bins(&boundary, 16);
    let bins = BinPolygons {
        bins: bin_polygons.iter()
            .enumerate()
            .map(|(id, polygon)| polygon_to_serializable(polygon, id))
            .collect(),
    };
    
    // Create area bins directory if it doesn't exist
    let area_bins_dir = paths.temp_dir.join("population").join("03_create_area_bins");
    fs::create_dir_all(&area_bins_dir)?;
    
    // Save to files
    fs::write(
        area_bins_dir.join("full-area.json"),
        serde_json::to_string_pretty(&full_area)?,
    )?;
    
    fs::write(
        area_bins_dir.join("bins.json"),
        serde_json::to_string_pretty(&bins)?,
    )?;
    
    Ok(persons)
}
