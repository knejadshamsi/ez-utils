use crossterm::{
    cursor::{Hide, MoveToColumn},
    execute,
    style::{Color, Print, ResetColor, SetForegroundColor},
    terminal::{Clear, ClearType},
};
use std::io::{Write, stdout};

fn print_base(step: u8, text: &str) {
    let mut stdout = stdout();
    execute!(
        stdout,
        Hide,
        Clear(ClearType::CurrentLine),
        MoveToColumn(0),
        Print(format!(
            "{}[STEP {}/4]{} {}",
            SetForegroundColor(Color::Magenta),
            step,
            ResetColor,
            text
        ))
    ).unwrap();
    stdout.flush().unwrap();
}

#[derive(Clone)]
pub struct VerbPair {
    active: String,
    completed: String,
}

pub struct ProcessProgress {
    step: u8,
    previous_parts: Vec<(VerbPair, String)>,
    pub current_part: (VerbPair, String),
}

impl ProcessProgress {
    pub fn new(step: u8, active_verb: &str, completed_verb: &str, target: &str) -> Self {
        let progress = Self {
            step,
            previous_parts: Vec::new(),
            current_part: (VerbPair {
                active: active_verb.to_string(),
                completed: completed_verb.to_string(),
            }, target.to_string()),
        };
        progress.print_progress();
        progress
    }

    pub fn update_target(&mut self, target: &str) {
        self.current_part.1 = target.to_string();
        self.print_progress();
    }

    pub fn advance(&mut self, active_verb: &str, completed_verb: &str, target: &str) {
        // Remove counter from current part before adding to previous parts
        let (base_target, _) = if let Some(idx) = self.current_part.1.find('[') {
            (&self.current_part.1[..idx-1], &self.current_part.1[idx..])
        } else {
            (self.current_part.1.as_str(), "")
        };
        
        let modified_part = (
            self.current_part.0.clone(),
            base_target.to_string()
        );
        self.previous_parts.push(modified_part);
        
        self.current_part = (VerbPair {
            active: active_verb.to_string(),
            completed: completed_verb.to_string(),
        }, target.to_string());
        self.print_progress();
    }

    pub fn print_progress(&self) {
        let mut text = String::new();

        for (verb_pair, target) in self.previous_parts.iter() {
            text.push_str(&format!("{}{}{} {}", 
                SetForegroundColor(Color::Green),
                &verb_pair.completed,
                ResetColor,
                target
            ));
            text.push_str(&format!(" {}→{} ", SetForegroundColor(Color::Cyan), ResetColor));
        }

        // Current part - verb in default color
        // Split out file names from the target
        let parts = self.current_part.1.split('[').collect::<Vec<_>>();
        if parts.len() > 1 {
            text.push_str(&format!("{}{} {}",
                ResetColor,
                &self.current_part.0.active,
                parts[0]
            ));
            // Format each filename with green color
            for part in parts.iter().skip(1) {
                if !part.is_empty() {
                    text.push_str(&format!("{}[{}{}",
                        SetForegroundColor(Color::Green),
                        part,
                        ResetColor
                    ));
                }
            }
        } else {
            text.push_str(&format!("{}{} {}",
                ResetColor,
                &self.current_part.0.active,
                &self.current_part.1,
            ));
        }

        print_base(self.step, &text);
    }

    pub fn complete(&self) {
        let mut stdout = stdout();
        execute!(
            stdout,
            Clear(ClearType::CurrentLine),
            MoveToColumn(0),
        ).unwrap();

        // Print completed step
        let mut text = String::new();
        for (i, (verb_pair, target)) in self.previous_parts.iter().chain(std::iter::once(&self.current_part)).enumerate() {
            text.push_str(&format!("{}{}{} ",
                SetForegroundColor(Color::Green),
                &verb_pair.completed,
                ResetColor
            ));

            // Split out file names from the target
            let parts = target.split('[').collect::<Vec<_>>();
            if parts.len() > 1 {
                text.push_str(parts[0]);
                // Format each filename with green color
                for part in parts.iter().skip(1) {
                    if !part.is_empty() {
                        text.push_str(&format!("{}[{}{}",
                            SetForegroundColor(Color::Green),
                            part,
                            ResetColor
                        ));
                    }
                }
            } else {
                text.push_str(target);
            }

            if i < self.previous_parts.len() {
                text.push_str(&format!(" {}→{} ", 
                    SetForegroundColor(Color::Cyan), 
                    ResetColor
                ));
            }
        }

        execute!(
            stdout,
            Print(format!(
                "{}[STEP {}/4]{} {}\n",
                SetForegroundColor(Color::Magenta),
                self.step,
                ResetColor,
                text
            ))
        ).unwrap();
        stdout.flush().unwrap();
    }
}

pub fn format_file_name(name: &str) -> String {
    format!("[{}]", name)
}
