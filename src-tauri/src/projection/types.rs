use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy)]
pub struct ProjectedCoord {
    pub lng: f64,
    pub lat: f64,
}

#[derive(Debug)]
pub struct NativeCoord {
    pub x: f64,
    pub y: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(tag = "kind", rename_all = "snake_case")]
pub enum CrsConfig {
    MontrealMtm8,
    TehranUtm39n,
    ParisLambert93,
    BerlinEtrs89Utm33n,
    LondonBng,
    NewYorkUtm18n,
    TokyoJgd2011Zone9,
    SydneyMga94Zone56,
    ZurichLv95,
    Custom {
        proj_string: String,
        center: [f64; 2],
        label: String,
    },
}

impl Default for CrsConfig {
    fn default() -> Self {
        CrsConfig::MontrealMtm8
    }
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct CrsInfo {
    pub kind: CrsConfig,
    pub label: String,
    pub center: [f64; 2],
}
