use crossterm::{
    cursor::{Hide, MoveToColumn},
    execute,
    style::{Color, Print, ResetColor, SetForegroundColor},
    terminal::{Clear, ClearType},
};
use std::io::{stdout, Write};

fn print_base(step: usize, total_steps: usize, text: &str) {
    let mut stdout = stdout();
    execute!(
        stdout,
        Hide,
        Clear(ClearType::CurrentLine),
        MoveToColumn(0),
        Print(format!(
            "{}[STEP {}/{}]{} {}",
            SetForegroundColor(Color::Magenta),
            step,
            total_steps,
            ResetColor,
            text
        ))
    ).unwrap();
    stdout.flush().unwrap();
}

pub fn print_step_progress(step: usize, total_steps: usize, current: usize, total: usize, msg: &str) {
    print_base(step, total_steps, &format!("{} {}[{}/{}]{}",
        msg,
        SetForegroundColor(Color::Yellow),
        current,
        total,
        ResetColor
    ));
}

pub fn print_step_start(step: usize, total_steps: usize, msg: &str, paths: Option<(&str, &str)>) {
    if let Some((index_path, chunks_path)) = paths {
        let green = SetForegroundColor(Color::Green);
        let reset = ResetColor;
        print_base(step, total_steps, &format!("{} {green}[{}{green}]{reset} {green}[{}{green}]{reset}", 
            msg,
            index_path,
            chunks_path
        ));
    } else {
        print_base(step, total_steps, msg);
    }
}

fn format_part(action: &str, target: &str) -> String {
    let capitalized = action.chars().next().map(|c| c.to_uppercase().collect::<String>())
        .unwrap_or_default() + &action[1..];
    format!("{}{}{} {}", 
        SetForegroundColor(Color::Green),
        capitalized,
        ResetColor,
        target
    )
}

pub fn print_step_without_counter(step: usize, total_steps: usize, msg: &str) {
    print_base(step, total_steps, msg);
}

pub fn print_step_complete(step: usize, total_steps: usize, base_msg: &str) {
    let mut stdout = stdout();

    let parts: Vec<_> = base_msg.split(" and ").collect();
    let formatted_msg = if parts.len() > 1 {
        let green = SetForegroundColor(Color::Green);
        let reset = ResetColor;
        format!("{}Validated{} necessary files: {green}[{}{green}]{reset} {green}[{}{green}]{reset}",
            green, reset,
            parts[0], parts[1])
    } else {
        format_part(&base_msg[..base_msg.find(' ').unwrap_or(0)], 
                   &base_msg[base_msg.find(' ').unwrap_or(0)..].trim())
    };

    execute!(
        stdout,
        Clear(ClearType::CurrentLine),
        MoveToColumn(0),
        Print(format!(
            "{}[STEP {}/{}]{} {}\n",
            SetForegroundColor(Color::Magenta),
            step,
            total_steps,
            ResetColor,
            formatted_msg
        ))
    ).unwrap();
    stdout.flush().unwrap();
}

pub fn print_split_progress(step: usize, total: usize, current_state: SplitState, progress: Option<(usize, usize)>) {
    let mut message = String::new();
    let green = SetForegroundColor(Color::Green);
    let cyan = SetForegroundColor(Color::Cyan);
    let yellow = SetForegroundColor(Color::Yellow);
    let reset = ResetColor;

    match current_state {
        SplitState::Finding => {
            if let Some((current, total)) = progress {
                message = format!("{}Finding{} xml separators {}[{}/{}]{}",
                    yellow, reset, yellow, current, total, reset);
            }
        }
        SplitState::FoundSeparators => {
            message = format!("{}Found{} xml separators", green, reset);
        }
        SplitState::Splitting(has_progress) => {
            message = format!("{}Found{} xml separators {}→{} ",
                green, reset, cyan, reset);
            if let Some((current, total)) = progress {
                message.push_str(&format!("{}Splitting{} network file {}[{}/{}]{}",
                    if has_progress { yellow } else { green }, reset,
                    yellow, current, total, reset));
            } else {
                message.push_str(&format!("{}Split{} network file", green, reset));
            }
        }
        SplitState::Chunking => {
            message = format!("{}Found{} xml separators {}→{} {}Split{} network file {}→{} ",
                green, reset, cyan, reset, green, reset, cyan, reset);
            if let Some((current, total)) = progress {
                message.push_str(&format!("{}Creating{} chunks {}[{}/{}]{}",
                    yellow, reset, yellow, current, total, reset));
            }
        }
        SplitState::Complete => {
            message = format!("{}Found{} xml separators {}→{} {}Split{} network file {}→{} {}Created{} chunks",
                green, reset, cyan, reset, green, reset, cyan, reset, green, reset);
        }
    }

    print_base(step, total, &message);
}

#[derive(Debug, Clone, Copy)]
pub enum SplitState {
    Finding,
    FoundSeparators,
    Splitting(bool),
    Chunking,
    Complete,
}
