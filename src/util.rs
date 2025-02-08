use anyhow::{Result, anyhow};
use std::env;
use std::path::PathBuf;
use std::process::Command;
use std::fs;

pub fn get_temp_dir() -> PathBuf {
    let temp_dir = std::env::current_dir()
        .unwrap_or_default()
        .join("temp");
    fs::create_dir_all(&temp_dir).unwrap_or_default();
    temp_dir
}

pub fn get_executable_path() -> Result<PathBuf> {
    eprintln!("Searching for ez-utils executable...");
    
    // First try to get the current executable path
    match env::current_exe() {
        Ok(exe_path) => {
            eprintln!("Current exe path: {}", exe_path.display());
            if exe_path.exists() {
                eprintln!("Found executable at current path");
                return Ok(exe_path);
            }
        }
        Err(e) => eprintln!("Failed to get current exe path: {}", e),
    }

    // Then try using the PATH environment
    match Command::new("where").arg("ez-utils").output() {
        Ok(output) => {
            if !output.stdout.is_empty() {
                let path_str = String::from_utf8_lossy(&output.stdout);
                let path = PathBuf::from(path_str.trim());
                eprintln!("Found in PATH: {}", path.display());
                if path.exists() {
                    eprintln!("Found executable in PATH");
                    return Ok(path);
                }
            }
        }
        Err(e) => eprintln!("Failed to run 'where' command: {}", e),
    }

    // Finally check cargo installation directory
    if let Ok(home) = env::var("USERPROFILE") {
        let cargo_bin = PathBuf::from(home).join(".cargo").join("bin").join("ez-utils.exe");
        eprintln!("Checking cargo bin: {}", cargo_bin.display());
        if cargo_bin.exists() {
            eprintln!("Found executable in cargo bin");
            return Ok(cargo_bin);
        }
    } else {
        eprintln!("Failed to get USERPROFILE");
    }

    let error_msg = anyhow!("ez-utils executable not found in PATH or cargo installation directory");
    eprintln!("Error: {}", error_msg);
    Err(error_msg)
}

pub fn ensure_executable_exists() -> Result<()> {
    eprintln!("Ensuring executable exists...");
    get_executable_path().map(|_| ())
}
