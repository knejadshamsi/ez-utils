use geo::Point;
use serde::{Serialize, Deserialize};
use std::path::PathBuf;
use std::collections::HashMap;
#[derive(Debug, Clone, Hash, Eq, PartialEq)]
pub struct LocationKey {
    pub x: i64,
    pub y: i64,
}

impl serde::Serialize for LocationKey {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        // Serialize as string in format "x,y"
        serializer.serialize_str(&format!("{},{}", self.x, self.y))
    }
}

impl LocationKey {
    pub fn from_coords(x: f64, y: f64) -> Self {
        // Convert to fixed-point representation with 6 decimal places
        let scale = 1_000_000;
        Self {
            x: (x * scale as f64) as i64,
            y: (y * scale as f64) as i64,
        }
    }
}

#[derive(Debug, Clone)]
pub struct PointWrapper(Point<f64>);

impl Serialize for PointWrapper {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut state = serializer.serialize_struct("Point", 2)?;
        state.serialize_field("x", &self.0.x())?;
        state.serialize_field("y", &self.0.y())?;
        state.end()
    }
}

impl<'de> Deserialize<'de> for PointWrapper {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        #[derive(Deserialize)]
        struct PointHelper {
            x: f64,
            y: f64,
        }

        let helper = PointHelper::deserialize(deserializer)?;
        Ok(PointWrapper(Point::new(helper.x, helper.y)))
    }
}

impl From<Point<f64>> for PointWrapper {
    fn from(point: Point<f64>) -> Self {
        PointWrapper(point)
    }
}

impl From<PointWrapper> for Point<f64> {
    fn from(wrapper: PointWrapper) -> Self {
        wrapper.0
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IndexEntry {
    pub id: String,
    #[serde(with = "point_serde")]
    pub coordinates: Point<f64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentIndex {
    pub agent_id: String,
    #[serde(with = "point_serde")]
    pub coordinates: Point<f64>,
    pub xml_content: String,
}

mod point_serde {
    use super::*;
    use serde::{Serializer, Deserializer};

    pub fn serialize<S>(point: &Point<f64>, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        PointWrapper::from(*point).serialize(serializer)
    }

    pub fn deserialize<'de, D>(deserializer: D) -> Result<Point<f64>, D::Error>
    where
        D: Deserializer<'de>,
    {
        Ok(PointWrapper::deserialize(deserializer)?.into())
    }
}

#[derive(Debug, Clone)]
#[allow(dead_code)]
pub struct Activity {
    pub id: String,
    pub activity_type: String,
    pub activity_order: usize,
    pub facility: Option<String>,
    pub start_time: Option<String>,
    pub end_time: Option<String>,
    pub coordinates: Option<Point<f64>>,
}

#[derive(Debug, Clone)]
pub struct Person {
    pub id: String,
    pub activities: Vec<Activity>,
}

#[derive(Debug, Clone)]
#[allow(dead_code)]
pub struct PopulationData {
    pub persons: Vec<Person>,
    pub current_scale: f64,
    pub total_population: usize,
    pub chunk_count: usize,
    pub last_chunk_size: usize,
    pub total_separator_comments: usize,
}

#[derive(Debug)]
#[allow(dead_code)]
pub struct ScaledPopulation {
    pub scale: f64,
    pub persons: Vec<Person>,
    pub location_density: HashMap<LocationKey, f64>,
}

#[derive(Debug)]
#[allow(dead_code)]
pub struct ProcessingPaths {
    pub base_dir: PathBuf,
    pub temp_dir: PathBuf,
    pub raw_dir: PathBuf,
    pub chunks_dir: PathBuf,
    pub index_dir: PathBuf,
    pub dense_dir: PathBuf,
    pub output_dir: PathBuf,
    pub scaled_chunks_dir: PathBuf,
    pub dense_new_dir: PathBuf,
    pub raw_chunks_dir: PathBuf,
    pub output_population_dir: PathBuf,
    pub assigned_dir: PathBuf,
}

#[derive(Debug)]
#[allow(dead_code)]
pub struct ChunkData {
    pub chunk_id: usize,
    pub agents: Vec<AgentIndex>,
    pub content: String,
    pub separator_comment_count: usize,
    pub chunk_number: usize,
}

#[derive(Debug, Clone)]
#[allow(dead_code)]
pub struct ScaleConfig {
    pub target_scale: f64,
    pub base_scale: f64,
    pub use_db: bool,
    pub delete_temp: bool,
    pub scales: Vec<f64>,
}
