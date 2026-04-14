pub(crate) mod types;

use proj4rs::{
    proj::Proj,
    transform::{transform, Transform, TransformClosure},
};
use tauri::State;

use crate::ez::types::{EzError, SessionManager};

pub use types::{CrsConfig, CrsInfo, NativeCoord, ProjectedCoord};

const MONTREAL_MTM8_PROJ: &str =
    "+proj=tmerc +lat_0=0 +lon_0=-73.5 +k=0.9999 +x_0=304800 +y_0=0 +datum=NAD83 +units=m +no_defs +type=crs";
const TEHRAN_UTM39N_PROJ: &str = "+proj=utm +zone=39 +datum=WGS84 +units=m +no_defs +type=crs";
const PARIS_LAMBERT93_PROJ: &str =
    "+proj=lcc +lat_0=46.5 +lon_0=3 +lat_1=49 +lat_2=44 +x_0=700000 +y_0=6600000 +ellps=GRS80 +towgs84=0,0,0,0,0,0,0 +units=m +no_defs +type=crs";
const BERLIN_ETRS89_UTM33N_PROJ: &str =
    "+proj=utm +zone=33 +ellps=GRS80 +towgs84=0,0,0,0,0,0,0 +units=m +no_defs +type=crs";
const LONDON_BNG_PROJ: &str =
    "+proj=tmerc +lat_0=49 +lon_0=-2 +k=0.9996012717 +x_0=400000 +y_0=-100000 +ellps=airy +towgs84=446.448,-125.157,542.06,0.15,0.247,0.842,-20.489 +units=m +no_defs +type=crs";
const NEW_YORK_UTM18N_PROJ: &str = "+proj=utm +zone=18 +datum=WGS84 +units=m +no_defs +type=crs";
const TOKYO_JGD2011_ZONE9_PROJ: &str =
    "+proj=tmerc +lat_0=36 +lon_0=139.833333333333 +k=0.9999 +x_0=0 +y_0=0 +ellps=GRS80 +towgs84=0,0,0,0,0,0,0 +units=m +no_defs +type=crs";
const SYDNEY_MGA94_ZONE56_PROJ: &str =
    "+proj=utm +zone=56 +south +ellps=GRS80 +towgs84=0,0,0,0,0,0,0 +units=m +no_defs +type=crs";
const ZURICH_LV95_PROJ: &str =
    "+proj=somerc +lat_0=46.9524055555556 +lon_0=7.43958333333333 +k_0=1 +x_0=2600000 +y_0=1200000 +ellps=bessel +towgs84=674.374,15.056,405.346,0,0,0,0 +units=m +no_defs +type=crs";
const WGS84_PROJ: &str = "+proj=longlat +datum=WGS84 +no_defs +type=crs";

pub(crate) fn crs_proj_string(config: &CrsConfig) -> &str {
    match config {
        CrsConfig::MontrealMtm8 => MONTREAL_MTM8_PROJ,
        CrsConfig::TehranUtm39n => TEHRAN_UTM39N_PROJ,
        CrsConfig::ParisLambert93 => PARIS_LAMBERT93_PROJ,
        CrsConfig::BerlinEtrs89Utm33n => BERLIN_ETRS89_UTM33N_PROJ,
        CrsConfig::LondonBng => LONDON_BNG_PROJ,
        CrsConfig::NewYorkUtm18n => NEW_YORK_UTM18N_PROJ,
        CrsConfig::TokyoJgd2011Zone9 => TOKYO_JGD2011_ZONE9_PROJ,
        CrsConfig::SydneyMga94Zone56 => SYDNEY_MGA94_ZONE56_PROJ,
        CrsConfig::ZurichLv95 => ZURICH_LV95_PROJ,
        CrsConfig::Custom { proj_string, .. } => proj_string,
    }
}

pub(crate) fn crs_center(config: &CrsConfig) -> [f64; 2] {
    match config {
        CrsConfig::MontrealMtm8 => [-73.5673, 45.5017],
        CrsConfig::TehranUtm39n => [51.3890, 35.6892],
        CrsConfig::ParisLambert93 => [2.3522, 48.8566],
        CrsConfig::BerlinEtrs89Utm33n => [13.4050, 52.5200],
        CrsConfig::LondonBng => [-0.1276, 51.5074],
        CrsConfig::NewYorkUtm18n => [-74.0060, 40.7128],
        CrsConfig::TokyoJgd2011Zone9 => [139.6917, 35.6895],
        CrsConfig::SydneyMga94Zone56 => [151.2093, -33.8688],
        CrsConfig::ZurichLv95 => [8.5417, 47.3769],
        CrsConfig::Custom { center, .. } => *center,
    }
}

pub(crate) fn crs_label(config: &CrsConfig) -> String {
    match config {
        CrsConfig::MontrealMtm8 => "Montreal, Canada".into(),
        CrsConfig::TehranUtm39n => "Tehran, Iran".into(),
        CrsConfig::ParisLambert93 => "Paris, France".into(),
        CrsConfig::BerlinEtrs89Utm33n => "Berlin, Germany".into(),
        CrsConfig::LondonBng => "London, UK".into(),
        CrsConfig::NewYorkUtm18n => "New York, USA".into(),
        CrsConfig::TokyoJgd2011Zone9 => "Tokyo, Japan".into(),
        CrsConfig::SydneyMga94Zone56 => "Sydney, Australia".into(),
        CrsConfig::ZurichLv95 => "Zurich, Switzerland".into(),
        CrsConfig::Custom { label, .. } => label.clone(),
    }
}

pub(crate) fn crs_info(config: &CrsConfig) -> CrsInfo {
    CrsInfo {
        kind: config.clone(),
        label: crs_label(config),
        center: crs_center(config),
    }
}

fn preset_list() -> Vec<CrsConfig> {
    vec![
        CrsConfig::MontrealMtm8,
        CrsConfig::TehranUtm39n,
        CrsConfig::ParisLambert93,
        CrsConfig::BerlinEtrs89Utm33n,
        CrsConfig::LondonBng,
        CrsConfig::NewYorkUtm18n,
        CrsConfig::TokyoJgd2011Zone9,
        CrsConfig::SydneyMga94Zone56,
        CrsConfig::ZurichLv95,
    ]
}

pub(crate) struct Projector {
    source: Proj,
    target: Proj,
}

impl Projector {
    pub(crate) fn new(config: &CrsConfig) -> Result<Self, EzError> {
        let source = Proj::from_proj_string(crs_proj_string(config)).map_err(|err| EzError::Io {
            message: format!("Failed to initialize source CRS: {err}"),
        })?;
        let target = Proj::from_proj_string(WGS84_PROJ).map_err(|err| EzError::Io {
            message: format!("Failed to initialize WGS84 projection: {err}"),
        })?;
        Ok(Self { source, target })
    }

    pub(crate) fn project_xy(&self, x: f64, y: f64) -> Result<ProjectedCoord, EzError> {
        let mut point = Point3 { x, y, z: 0.0 };
        transform(&self.source, &self.target, &mut point).map_err(|err| EzError::Io {
            message: format!("Failed to project coordinate to WGS84: {err}"),
        })?;
        Ok(ProjectedCoord {
            lng: point.x.to_degrees(),
            lat: point.y.to_degrees(),
        })
    }

    pub(crate) fn unproject_lng_lat(&self, lng: f64, lat: f64) -> Result<NativeCoord, EzError> {
        let mut point = Point3 {
            x: lng.to_radians(),
            y: lat.to_radians(),
            z: 0.0,
        };
        transform(&self.target, &self.source, &mut point).map_err(|err| EzError::Io {
            message: format!("Failed to project WGS84 coordinate to source CRS: {err}"),
        })?;
        Ok(NativeCoord {
            x: point.x,
            y: point.y,
        })
    }
}

struct Point3 {
    x: f64,
    y: f64,
    z: f64,
}

impl Transform for Point3 {
    fn transform_coordinates<F: TransformClosure>(
        &mut self,
        f: &mut F,
    ) -> proj4rs::errors::Result<()> {
        f(self.x, self.y, self.z).map(|(x, y, z)| {
            self.x = x;
            self.y = y;
            self.z = z;
        })
    }
}

#[tauri::command]
pub fn list_crs_presets() -> Vec<CrsInfo> {
    preset_list().iter().map(crs_info).collect()
}

#[tauri::command]
pub fn set_crs(
    config: CrsConfig,
    manager: State<'_, SessionManager>,
) -> Result<[f64; 2], EzError> {
    let mut guard = manager.session.lock().map_err(|_| EzError::SessionBusy)?;
    let session = guard.as_mut().ok_or(EzError::NoActiveSession)?;
    if !session.ui_state.sources.is_empty() {
        return Err(EzError::CrsLockedWithSources);
    }
    let center = crs_center(&config);
    session.ui_state.settings.crs = config;
    session.dirty = true;
    Ok(center)
}
