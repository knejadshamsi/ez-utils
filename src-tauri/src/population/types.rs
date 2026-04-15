use serde::{Deserialize, Serialize};

#[derive(Debug)]
pub(crate) struct PersonState {
    pub person_id: String,
    pub next_plan_index: i64,
    pub wrote_attributes: bool,
}

#[derive(Debug)]
pub(crate) struct PersonAttributesRecord {
    pub person_id: String,
    pub attributes_blob: Option<String>,
}

#[derive(Debug)]
pub(crate) struct PlanRecord {
    pub person_id: String,
    pub plan_id: String,
    pub plan_index: i64,
    pub selected: i64,
    pub plan_blob: String,
}

#[derive(Debug)]
pub(crate) struct ActivityCoordRecord {
    pub person_id: String,
    pub plan_id: String,
    pub plan_index: i64,
    pub activity_index: i64,
    pub lng: f64,
    pub lat: f64,
}

#[derive(Debug)]
pub(crate) struct RawActivityCoord {
    pub activity_index: i64,
    pub x: f64,
    pub y: f64,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ActivityCoordPayload {
    pub person_id: String,
    pub plan_id: String,
    pub plan_index: i64,
    pub activity_index: i64,
    pub lng: f64,
    pub lat: f64,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct QueryPopulationBboxPayload {
    pub rows: Vec<ActivityCoordPayload>,
    pub total_rows: i64,
    pub total_people: i64,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct PersonPlanPayload {
    pub plan_id: String,
    pub plan_index: i64,
    pub selected: i64,
    pub plan_blob: String,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct PersonPayload {
    pub person_id: String,
    pub attributes_blob: Option<String>,
    pub plans: Vec<PersonPlanPayload>,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct UpdatePlanCoordInput {
    pub activity_index: i64,
    pub lng: f64,
    pub lat: f64,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct PlanActivityEditInput {
    pub original_index: Option<i64>,
    pub type_name: String,
    pub start_time: Option<String>,
    pub end_time: Option<String>,
    pub lng: Option<f64>,
    pub lat: Option<f64>,
    pub dirty: bool,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct PlanLegEditInput {
    pub original_index: Option<i64>,
    pub mode: String,
    pub travel_time: Option<String>,
    pub dirty: bool,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ApplyPlanEditsInput {
    pub activities: Vec<PlanActivityEditInput>,
    pub legs: Vec<PlanLegEditInput>,
}

