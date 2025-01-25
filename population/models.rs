use geo::Point;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Activity {
    pub id: String,
    pub activity_order: i32,
    pub activity_type: String,
    pub facility: Option<String>,
    pub start_time: Option<String>,
    pub end_time: Option<String>,
    pub coordinates: Option<Point<f64>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Person {
    pub id: String,
    pub activities: Vec<Activity>,
}

#[derive(Debug)]
pub struct PopulationData {
    pub persons: Vec<Person>,
    pub total_population: i32,
    pub current_scale: f64,
}

#[derive(Debug)]
pub struct ScaledPopulation {
    pub scale: f64,
    pub persons: Vec<Person>,
    pub location_density: HashMap<(f64, f64), f64>,
}
