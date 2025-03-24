pub mod cli;
pub mod core;
pub mod steps;
pub mod utils;

pub use core::{
    models::{
        PtError,
        ServiceDay,
        ServicePattern,
        ServiceRouteMapping,
        SelectedServices,
        TripMapping,
        StopSequence,
        StopInfo,
        StopDetails,
    },
    setup::run,
};

pub use steps::{
    validate_gtfs_files,
    handle_service_selection,
    handle_trip_collection,
    handle_schedule_building,
};
