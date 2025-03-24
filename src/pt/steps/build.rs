use std::collections::HashMap;
use std::fs;
use std::io::{Cursor, Write};
use std::path::Path;
use quick_xml::events::{Event, BytesStart};
use quick_xml::Writer;
use serde_json;

use crate::pt::core::models::{PtError, StopDetails, StopSequence};
use crate::pt::utils::display::ProcessProgress;

struct IndentedWriter {
    writer: Writer<Cursor<Vec<u8>>>,
    indent_level: usize,
    indent_str: String,
}

impl IndentedWriter {
    fn new() -> Self {
        Self {
            writer: Writer::new(Cursor::new(Vec::new())),
            indent_level: 0,
            indent_str: "    ".to_string(),
        }
    }

    fn write_event(&mut self, event: Event) -> Result<(), PtError> {
        match &event {
            Event::Start(_) => {
                self.write_indentation()?;
                self.writer.write_event(event)?;
                self.write_newline()?;
                self.indent_level += 1;
            }
            Event::End(_) => {
                self.indent_level -= 1;
                self.write_indentation()?;
                self.writer.write_event(event)?;
                self.write_newline()?;
            }
            _ => self.writer.write_event(event)?,
        }
        Ok(())
    }

    fn write_indentation(&mut self) -> Result<(), PtError> {
        let indent = self.indent_str.repeat(self.indent_level);
        self.writer.get_mut().write_all(indent.as_bytes())?;
        Ok(())
    }

    fn write_newline(&mut self) -> Result<(), PtError> {
        self.writer.get_mut().write_all(b"\n")?;
        Ok(())
    }

    fn into_string(self) -> Result<String, PtError> {
        Ok(String::from_utf8(self.writer.into_inner().into_inner())?)
    }
}

pub fn handle_schedule_building(output_dir: &Path) -> Result<(), PtError> {
    let mut progress = ProcessProgress::new(4, "Processing", "Processed", "transit schedule");

    let mut writer = IndentedWriter::new();
    let schedule_elem = BytesStart::new("transitSchedule");
    writer.write_event(Event::Start(schedule_elem.clone()))?;

    progress.advance("Processing", "Processed", "stops");
    process_stops(&mut writer)?;

    progress.advance("Processing", "Processed", "routes");
    process_routes(&mut writer, &mut progress)?;

    writer.write_event(Event::End(schedule_elem.to_end()))?;

    progress.advance("Saving", "Saved", "schedule");
    let pt_dir = output_dir.join("pt");
    fs::create_dir_all(&pt_dir)?;
    
    let final_content = writer.into_string()?;
    fs::write(
        pt_dir.join("transitSchedule.xml"),
        final_content,
    )?;

    progress.complete();
    Ok(())
}

fn process_stops(writer: &mut IndentedWriter) -> Result<(), PtError> {
    let stop_details: HashMap<String, StopDetails> = serde_json::from_str(
        &fs::read_to_string(Path::new("temp/pt/03-trips/stop_details.json"))?,
    ).map_err(|e| PtError::JsonError(e.to_string()))?;

    let stops_elem = BytesStart::new("transitStops");
    writer.write_event(Event::Start(stops_elem.clone()))?;
    
    for (_stop_id, details) in stop_details.iter() {
        let mut stop_elem = BytesStart::new("stopFacility");
        stop_elem.push_attribute(("id", details.stop_id.as_str()));
        stop_elem.push_attribute(("x", details.stop_lon.to_string().as_str()));
        stop_elem.push_attribute(("y", details.stop_lat.to_string().as_str()));
        stop_elem.push_attribute(("name", details.stop_name.as_str()));
        
        writer.write_event(Event::Start(stop_elem.borrow()))?;
        writer.write_event(Event::End(stop_elem.to_end()))?;
    }

    writer.write_event(Event::End(stops_elem.to_end()))?;
    Ok(())
}

fn process_routes(writer: &mut IndentedWriter, progress: &mut ProcessProgress) -> Result<(), PtError> {
    let stop_sequences_dir = Path::new("temp/pt/03-trips/stop_sequences");
    
    let transit_lines = BytesStart::new("transitLines");
    writer.write_event(Event::Start(transit_lines.clone()))?;

    let mut sequences = Vec::new();
    for entry in fs::read_dir(stop_sequences_dir)? {
        let entry = entry?;
        if entry.path().extension().unwrap_or_default() == "json" {
            let sequence: StopSequence = serde_json::from_str(
                &fs::read_to_string(entry.path())?,
            ).map_err(|e| PtError::JsonError(e.to_string()))?;
            sequences.push(sequence);
        }
    }

    let total = sequences.len();
    let mut processed = 0;
    for sequence in sequences {
        let mut route_elem = BytesStart::new("transitRoute");
        processed += 1;
        progress.update_target(&format!("routes [{}/{}]", processed, total));
        route_elem.push_attribute(("id", sequence.trip_id.as_str()));
        writer.write_event(Event::Start(route_elem.clone()))?;

        // Get first departure time
        let first_departure_time = sequence.stops[0].departure_time.clone();

        // Add route stops
        let stops_elem = BytesStart::new("routeProfile");
        writer.write_event(Event::Start(stops_elem.clone()))?;
        
        for stop in &sequence.stops {
            let mut stop_elem = BytesStart::new("stop");
            stop_elem.push_attribute(("refId", stop.stop_id.as_str()));
            stop_elem.push_attribute(("arrivalOffset", stop.arrival_time.as_str()));
            stop_elem.push_attribute(("departureOffset", stop.departure_time.as_str()));
            writer.write_event(Event::Start(stop_elem.clone()))?;
            writer.write_event(Event::End(stop_elem.to_end()))?;
        }
        writer.write_event(Event::End(stops_elem.to_end()))?;

        // Add departures
        let departures_elem = BytesStart::new("departures");
        writer.write_event(Event::Start(departures_elem.clone()))?;
        
        let mut departure_elem = BytesStart::new("departure");
        departure_elem.push_attribute(("id", "1"));
        departure_elem.push_attribute(("departureTime", first_departure_time.as_str()));
        writer.write_event(Event::Start(departure_elem.clone()))?;
        writer.write_event(Event::End(departure_elem.to_end()))?;
        
        writer.write_event(Event::End(departures_elem.to_end()))?;
        writer.write_event(Event::End(route_elem.to_end()))?;
    }

    writer.write_event(Event::End(transit_lines.to_end()))?;
    Ok(())
}
