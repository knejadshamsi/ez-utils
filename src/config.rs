use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;

#[derive(Debug, Serialize, Deserialize)]
pub struct Settings {
    pub divider: String,
    pub chunk_size: usize,
    pub total_population: i32,
    pub area_size: f64,
}

impl Default for Settings {
    fn default() -> Self {
        Self {
            divider: "</person>".to_string(),
            chunk_size: 1000, // Number of persons per chunk
            total_population: 1000000,
            area_size: 1000.0,
        }
    }
}

impl Settings {
    pub fn load() -> Result<Self> {
        let config_path = get_config_path()?;
        
        if !config_path.exists() {
            let settings = Self::default();
            settings.save()?;
            return Ok(settings);
        }

        let content = fs::read_to_string(&config_path)
            .context("Failed to read config file")?;
        
        toml::from_str(&content)
            .context("Failed to parse config file")
    }

    pub fn save(&self) -> Result<()> {
        let config_path = get_config_path()?;
        
        if let Some(parent) = config_path.parent() {
            fs::create_dir_all(parent)
                .context("Failed to create config directory")?;
        }

        let content = toml::to_string(self)
            .context("Failed to serialize settings")?;
        
        fs::write(&config_path, content)
            .context("Failed to write config file")?;
        
        Ok(())
    }

    pub fn set(&mut self, key: &str, value: &str) -> Result<()> {
        match key {
            "divider" => self.divider = value.to_string(),
            "chunk_size" => self.chunk_size = value.parse()
                .context("Invalid chunk size")?,
            "total_population" => self.total_population = value.parse()
                .context("Invalid total population")?,
            "area_size" => self.area_size = value.parse()
                .context("Invalid area size")?,
            _ => anyhow::bail!("Unknown setting: {}", key),
        }
        self.save()
    }
}

fn get_config_path() -> Result<PathBuf> {
    let config_dir = dirs::config_dir()
        .context("Could not determine config directory")?
        .join("ez-utils");
    
    Ok(config_dir.join("config.toml"))
}
