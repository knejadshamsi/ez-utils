use anyhow::Result;
use crossterm::{
    cursor::{Hide, MoveToColumn, Show},
    execute,
    style::{Color, Print, ResetColor, SetForegroundColor},
    terminal::{Clear, ClearType},
};
use std::io::{stdout, Write};

pub fn print_step(step: usize, total_steps: usize, chunks: usize, population: usize) {
    let mut stdout = stdout();
    execute!(
        stdout,
        Hide,
        Clear(ClearType::CurrentLine),
        MoveToColumn(0),
        Print(format!(
            "{}[STEP {}/{}]{} Creating {}{}{} smaller chunks (Person count: {}{}{})",
            SetForegroundColor(Color::Magenta),
            step,
            total_steps,
            ResetColor,
            SetForegroundColor(Color::Yellow),
            chunks,
            ResetColor,
            SetForegroundColor(Color::Yellow),
            population,
            ResetColor
        ))
    ).unwrap();
    stdout.flush().unwrap();
}

pub fn step_success(step: usize, total_steps: usize, chunks: usize, population: usize) {
    let mut stdout = stdout();
    execute!(
        stdout,
        Clear(ClearType::CurrentLine),
        MoveToColumn(0),
        Print(format!(
            "{}[STEP {}/{}]{} {}Created{} {}{}{}{}smaller chunks (Person count: {}{}{}) {}[Chunks fixed]{}\n",
            SetForegroundColor(Color::Magenta),
            step,
            total_steps,
            ResetColor,
            SetForegroundColor(Color::Green),
            ResetColor,
            SetForegroundColor(Color::Yellow),
            chunks,
            ResetColor,
            " ",
            SetForegroundColor(Color::Yellow),
            population,
            ResetColor,
            SetForegroundColor(Color::Green),
            ResetColor
        ))
    ).unwrap();
}

pub struct ProcessProgress {
    pub step: usize,
    pub total_steps: usize,
}

impl ProcessProgress {
    pub fn new(step: usize, total_steps: usize, _total: u64) -> Result<Self> {
        let mut stdout = stdout();
        execute!(stdout, Hide).unwrap();
        Ok(Self {
            step,
            total_steps,
        })
    }

    pub fn print_step_3_establishing(&mut self) {
        self.print_base("establishing zone");
    }

    pub fn print_step_3_established(&mut self) {
        self.print_line(
            &[("established", "zone")],
            &[("partitioning", "zone into bins")]
        );
    }

    pub fn print_step_3_complete(&mut self) {
        self.print_complete(&[("established", "zone"), ("partitioned", "zone into bins")]);
    }

    pub fn print_step_4_progress(&mut self) {
        self.print_base("assigning person to bins");
    }

    pub fn print_step_4_complete(&mut self) {
        self.print_complete(&[("assigned", "person to bins")]);
    }

    pub fn print_step_5_calculating(&mut self) {
        self.print_base("calculating ratio");
    }

    pub fn print_step_5_calculated(&mut self) {
        self.print_line(
            &[("calculated", "ratio")],
            &[("scaling", "bins")]
        );
    }

    pub fn print_step_5_scaled(&mut self) {
        self.print_line(
            &[("calculated", "ratio"), ("scaled", "bins")],
            &[("distributing", "agents")]
        );
    }

    pub fn print_step_5_complete(&mut self) {
        self.print_complete(&[("calculated", "ratio"), ("scaled", "bins"), ("distributed", "agents")]);
    }

    pub fn print_step_6_indexing(&mut self, percentage: u32) -> Result<()> {
        self.print_base(&format!("indexing representation: {}{}%{}", 
            SetForegroundColor(Color::Yellow), percentage, ResetColor));
        Ok(())
    }

    pub fn print_step_6_indexed(&mut self, current: u64, total: u64) -> Result<()> {
        self.print_line(
            &[("indexed", "representation")],
            &[("creating", &format!("chunks {}[{}/{}]{}",
                SetForegroundColor(Color::Yellow),
                current,
                total,
                ResetColor))]
        );
        Ok(())
    }

    pub fn print_step_6_created(&mut self, percentage: u32) -> Result<()> {
        self.print_line(
            &[("indexed", "representation"), ("created", "chunks")],
            &[("combining", &format!("data: {}{}%{}", 
                SetForegroundColor(Color::Yellow), percentage, ResetColor))]
        );
        Ok(())
    }

    pub fn print_step_6_complete(&mut self) {
        self.print_complete(&[("indexed", "representation"), ("created", "chunks"), ("combined", "data")]);
    }

    pub fn update_progress(&mut self, current: u64, total: u64, _title: &str) {
        if self.step == 2 {
            if current == total {
                self.print_complete(&[("indexed", "chunks")]);
            } else {
                self.print_base(&format!("indexed chunks {}[{}/{}]{}",
                    SetForegroundColor(Color::Yellow),
                    current,
                    total,
                    ResetColor));
            }
        }
    }

    pub fn finish(&mut self) {
        let mut stdout = stdout();
        execute!(stdout, Show).unwrap();
    }

    fn format_part(&self, action: &str, target: &str) -> String {
        let capitalized = action.chars().next().map(|c| c.to_uppercase().collect::<String>())
            .unwrap_or_default() + &action[1..];
        format!("{}{}{} {}", 
            SetForegroundColor(Color::Green),
            capitalized,
            ResetColor,
            target
        )
    }

    fn print_base(&mut self, text: &str) {
        let mut stdout = stdout();
        execute!(
            stdout,
            Clear(ClearType::CurrentLine),
            MoveToColumn(0),
            Print(format!(
                "{}[STEP {}/{}]{} {}",
                SetForegroundColor(Color::Magenta),
                self.step,
                self.total_steps,
                ResetColor,
                text
            ))
        ).unwrap();
        stdout.flush().unwrap();
    }

    fn print_complete(&mut self, parts: &[(&str, &str)]) {
        let mut text = format!(
            "{}[STEP {}/{}]{} ",
            SetForegroundColor(Color::Magenta),
            self.step,
            self.total_steps,
            ResetColor
        );

        for (i, (action, target)) in parts.iter().enumerate() {
            text.push_str(&self.format_part(action, target));
            if i < parts.len() - 1 {
                text.push_str(&format!(" {}→{} ", SetForegroundColor(Color::Cyan), ResetColor));
            }
        }
        text.push('\n');

        let mut stdout = stdout();
        execute!(
            stdout,
            Clear(ClearType::CurrentLine),
            MoveToColumn(0),
            Print(text)
        ).unwrap();
        stdout.flush().unwrap();
    }

    fn print_line(&mut self, done_parts: &[(&str, &str)], current_parts: &[(&str, &str)]) {
        let mut text = format!(
            "{}[STEP {}/{}]{} ",
            SetForegroundColor(Color::Magenta),
            self.step,
            self.total_steps,
            ResetColor
        );

        // Add completed parts
        for (i, (action, target)) in done_parts.iter().enumerate() {
            text.push_str(&self.format_part(action, target));
            if i < done_parts.len() - 1 || !current_parts.is_empty() {
                text.push_str(&format!(" {}→{} ", SetForegroundColor(Color::Cyan), ResetColor));
            }
        }

        // Add current parts
        for (i, (action, target)) in current_parts.iter().enumerate() {
            text.push_str(&self.format_part(action, target));
            if i < current_parts.len() - 1 {
                text.push_str(&format!(" {}→{} ", SetForegroundColor(Color::Cyan), ResetColor));
            }
        }

        let mut stdout = stdout();
        execute!(
            stdout,
            Clear(ClearType::CurrentLine),
            MoveToColumn(0),
            Print(text)
        ).unwrap();
        stdout.flush().unwrap();
    }
}
