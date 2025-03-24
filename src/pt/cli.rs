use std::path::PathBuf;
use clap::{Parser, Args};

use crate::pt::core::{models::ServiceDay, setup, PtError};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
pub struct Cli {
    /// Path to the GTFS directory
    #[arg()]
    pub input: PathBuf,

    #[command(flatten)]
    pub day_options: DayOptions,

    /// Clean up temporary files after processing
    #[arg(short, long)]
    pub clean: bool,
}

#[derive(Args, Debug, Default)]
#[group(required = false, multiple = false)]
pub struct DayOptions {
    /// Select from weekend services
    #[arg(short = 'w', long)]
    pub weekend: bool,

    /// Select from holiday services
    #[arg(short = 'h', long)]
    pub holiday: bool,

    /// Select from Monday services
    #[arg(long = "monday", short = 'm')]
    pub monday: bool,

    /// Select from Tuesday services
    #[arg(long = "tuesday", short = 't')]
    pub tuesday: bool,

    /// Select from Wednesday services
    #[arg(long = "wednesday", short = 'e')]
    pub wednesday: bool,

    /// Select from Thursday services
    #[arg(long = "thursday", short = 'r')]
    pub thursday: bool,

    /// Select from Friday services
    #[arg(long = "friday", short = 'f')]
    pub friday: bool,
}

impl DayOptions {
    pub fn validate(&self) -> Result<(), &'static str> {
        let flags = [
            self.weekend,
            self.holiday,
            self.monday,
            self.tuesday,
            self.wednesday,
            self.thursday,
            self.friday,
        ];
        
        if flags.iter().filter(|&&x| x).count() > 1 {
            return Err("Only one day flag can be specified");
        }
        
        Ok(())
    }

    pub fn get_service_day(&self) -> ServiceDay {
        match true {
            _ if self.weekend => ServiceDay::Weekend,
            _ if self.holiday => ServiceDay::Holiday,
            _ if self.monday => ServiceDay::Monday,
            _ if self.tuesday => ServiceDay::Tuesday,
            _ if self.wednesday => ServiceDay::Wednesday,
            _ if self.thursday => ServiceDay::Thursday,
            _ if self.friday => ServiceDay::Friday,
            _ => ServiceDay::Weekday, // Default to weekday if no option selected
        }
    }
}

pub fn run() -> Result<(), PtError> {
    let cli = Cli::parse();
    setup::run(cli)
}
