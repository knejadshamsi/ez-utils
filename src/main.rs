use anyhow::Result;
use std::{fs, path::PathBuf};
use tokio;

use ez_utils::{
    cli::{self, Commands},
    network, population, pt, settings,
    config::Settings,
    population::{
        core::{
            config::Database,
        },
        utils::file::setup_processing_directories,
    },
};

async fn process_population(
    input: &PathBuf,
    scales: &Option<Vec<f64>>,
    db: Option<&Database>,
    delete: bool,
) -> Result<()> {
    let _settings = Settings::load()?;
    let paths = setup_processing_directories(&input)?;
    
    // Process population using 6-step process
    population::steps::process::process_population_file(
        input,
        &paths,
        &population::core::models::ScaleConfig {
            target_scale: 10.0, // Not used for bin scaling
            base_scale: 100.0,
            use_db: db.is_some(),
            delete_temp: delete,
            scales: scales.clone().unwrap_or_default(), // Empty vec triggers default 1-10 scales
        }
    ).await?;

    if delete {
        fs::remove_dir_all(&paths.temp_dir)?;
    }
    Ok(())
}

#[tokio::main]
async fn main() -> Result<()> {
    let cli = cli::parse_cli()?;

    match cli.command {
        Commands::Settings(cmd) => {
            settings::handle_command(cmd).await?;
        }
        Commands::Population { input, scales, db, delete } => {
            let scales = scales.map(|s| {
                s.split(',')
                    .filter_map(|scale| scale.trim().parse::<f64>().ok())
                    .collect::<Vec<_>>()
            });

            let db = if db {
                let db = Database::new().await?;
                db.ensure_tables_exist().await?;
                Some(db)
            } else {
                None
            };

            process_population(&input, &scales, db.as_ref(), delete).await?;
        },
        Commands::Network { input: _, db: _, delete: _ } => {
            network::cli::run().await?;
        },
        Commands::Pt { input: _, modes: _, day: _, delete: _ } => {
            pt::cli::run().map_err(|e| anyhow::anyhow!("PT CLI error: {}", e))?;
        },
    }

    Ok(())
}
