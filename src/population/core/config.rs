use anyhow::{Context, Result};
use sqlx::{PgPool, postgres::PgPoolOptions};
use std::env;

pub struct Database {
    pool: PgPool,
}

impl Database {
    pub async fn new() -> Result<Self> {
        let host = env::var("HOST_NAME").context("HOST_NAME not set")?;
        let database = env::var("DATA_BASE").context("DATA_BASE not set")?;
        let user = env::var("USER").context("USER not set")?;
        let password = env::var("PASS").context("PASS not set")?;

        let database_url = format!(
            "postgres://{}:{}@{}/{}",
            user, password, host, database
        );

        let pool = PgPoolOptions::new()
            .max_connections(5)
            .connect(&database_url)
            .await
            .context("Failed to create database pool")?;

        Ok(Self { pool })
    }

    pub async fn ensure_tables_exist(&self) -> Result<()> {
        sqlx::query(
            r#"
            DROP TABLE IF EXISTS agents;
            
            CREATE TABLE agents (
                agent_id TEXT PRIMARY KEY,
                start_coord GEOMETRY(Point, 4326),
                percentages SMALLINT[] NOT NULL,
                agent_xml TEXT
            );

            CREATE INDEX IF NOT EXISTS agents_start_coord_idx 
            ON agents USING GIST(start_coord);
            "#,
        )
        .execute(&self.pool)
        .await
        .context("Failed to create tables")?;

        Ok(())
    }

    pub async fn insert_agent(
        &self,
        agent_id: &str,
        start_coord: &str,
        percentages: &[i32],
        agent_xml: &str,
    ) -> Result<()> {
        sqlx::query(
            r#"
            INSERT INTO agents (agent_id, start_coord, percentages, agent_xml)
            VALUES ($1, ST_GeomFromText($2, 4326), $3, $4)
            ON CONFLICT (agent_id) DO UPDATE
            SET start_coord = EXCLUDED.start_coord,
                percentages = EXCLUDED.percentages,
                agent_xml = EXCLUDED.agent_xml
            "#,
        )
        .bind(agent_id)
        .bind(start_coord)
        .bind(percentages)
        .bind(agent_xml)
        .execute(&self.pool)
        .await
        .context("Failed to insert agent")?;

        Ok(())
    }
}
