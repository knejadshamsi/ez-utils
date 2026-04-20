use std::{path::PathBuf, sync::Mutex};

use serde::{Deserialize, Serialize};

use crate::projection::CrsConfig;

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct NetworkMetadata {
    pub network_attributes_blob: Option<String>,
    pub links_tag_blob: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct TransitMetadata {
    pub transit_schedule_attributes_blob: Option<String>,
    pub transit_stops_tag_blob: String,
    #[serde(default)]
    pub transit_minimal_transfers_tag_blob: String,
    pub transit_lines_tag_blob: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub vehicles: Option<TransitVehiclesMetadata>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub linked_network_source_name: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct TransitVehiclesMetadata {
    pub root_tag_attributes_blob: String,
    pub source_file_name: String,
}

pub(crate) const DEFAULT_MAP_CENTER: [f64; 2] = [-73.5673, 45.5017];
pub(crate) const DEFAULT_MAP_ZOOM: u8 = 12;
pub(crate) const DEFAULT_AUTOSAVE_INTERVAL_MINUTES: u32 = 5;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum SourceKind {
    Network,
    Population,
    Transit,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SourceEntry {
    pub id: String,
    pub name: String,
    pub kind: SourceKind,
    pub color: String,
    pub opacity: f64,
    pub visible: bool,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub network_metadata: Option<NetworkMetadata>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub transit_metadata: Option<TransitMetadata>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AutosaveSettings {
    pub enabled: bool,
    pub interval_minutes: u32,
}

impl Default for AutosaveSettings {
    fn default() -> Self {
        Self {
            enabled: false,
            interval_minutes: DEFAULT_AUTOSAVE_INTERVAL_MINUTES,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Settings {
    pub theme: String,
    pub locale: String,
    #[serde(default)]
    pub autosave_settings: AutosaveSettings,
    #[serde(default)]
    pub crs: CrsConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct MapState {
    pub center: [f64; 2],
    pub zoom: u8,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct UiState {
    pub version: u32,
    pub sources: Vec<SourceEntry>,
    pub settings: Settings,
    pub source_popover_open: bool,
    pub open_drawers: Vec<String>,
    pub map_state: MapState,
}

impl Default for UiState {
    fn default() -> Self {
        Self {
            version: 1,
            sources: Vec::new(),
            settings: Settings {
                theme: "dark".into(),
                locale: "en".into(),
                autosave_settings: AutosaveSettings::default(),
                crs: CrsConfig::default(),
            },
            source_popover_open: false,
            open_drawers: Vec::new(),
            map_state: MapState {
                center: DEFAULT_MAP_CENTER,
                zoom: DEFAULT_MAP_ZOOM,
            },
        }
    }
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct StatePayload {
    pub ez_path: String,
    pub dirty: bool,
    pub ui_state: UiState,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ImportedSourcePayload {
    pub name: String,
    pub kind: SourceKind,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub network_metadata: Option<NetworkMetadata>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub transit_metadata: Option<TransitMetadata>,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct NewSessionPayload {
    pub state: StatePayload,
    pub imported: ImportedSourcePayload,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum CloseMode {
    Save,
    Discard,
}

#[derive(Debug, Serialize)]
#[serde(tag = "kind", rename_all = "snake_case")]
pub enum EzError {
    NoActiveSession,
    SessionBusy,
    ArchiveMissing,
    ArchiveCorrupt,
    UnsupportedUiVersion { version: u32 },
    UnwritableLocation { path: String },
    SourceNameInvalid { message: String },
    SourceNotFound { name: String },
    ImportNotImplemented { source_kind: String },
    SourceLimitReached { limit: usize },
    UnsupportedSource { message: String },
    PopulationPersonNotFound { person_id: String },
    PopulationPlanNotFound { person_id: String, plan_id: String },
    PopulationLastPlan { person_id: String },
    NetworkNodeNotFound { node_id: String },
    NetworkLinkNotFound { link_id: String },
    NetworkDuplicateNodeId { node_id: String },
    NetworkDuplicateLinkId { link_id: String },
    TransitLineNotFound { line_id: String },
    TransitDuplicateLineId { line_id: String },
    TransitLineIdInvalid { message: String },
    TransitDuplicateRouteId { line_id: String, route_id: String },
    TransitRouteIdInvalid { message: String },
    TransitRouteNotFound { line_id: String, route_id: String },
    TransitTransportModeInvalid { message: String },
    TransitStopFacilityNotFound { stop_id: String },
    TransitDuplicateStopFacilityId { stop_id: String },
    TransitStopFacilityIdInvalid { message: String },
    TransitProfileStopInvalid { message: String },
    CrsLockedWithSources,
    Io { message: String },
}

impl From<String> for EzError {
    fn from(message: String) -> Self {
        EzError::Io { message }
    }
}

pub(crate) struct Session {
    pub ez_path: PathBuf,
    pub work_dir: PathBuf,
    pub dirty: bool,
    pub ui_state: UiState,
}

pub struct SessionManager {
    pub(crate) session: Mutex<Option<Session>>,
}

impl Default for SessionManager {
    fn default() -> Self {
        Self {
            session: Mutex::new(None),
        }
    }
}

pub(crate) fn session_payload(session: &Session) -> StatePayload {
    StatePayload {
        ez_path: session.ez_path.to_string_lossy().to_string(),
        dirty: session.dirty,
        ui_state: session.ui_state.clone(),
    }
}
