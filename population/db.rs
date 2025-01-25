use anyhow::{Context, Result};
use sqlx::{PgPool, postgres::PgPoolOptions};
use std::env;
use crate::population::models::{Person, Activity};

pub struct Database {
    pool: PgPool,
}

impl Database {
    pub async fn new() -> Result<Self> {
        let database_url = env::var("DATABASE_URL")
            .context("DATABASE_URL must be set")?;

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
            CREATE TABLE IF NOT EXISTS agents (
                id TEXT PRIMARY KEY,
                start_coordinates GEOMETRY(Point, 4326),
                percentages INTEGER[],
                agent_xml TEXT
            );

            CREATE TABLE IF NOT EXISTS link_coordinates (
                link_id TEXT PRIMARY KEY,
                from_node GEOMETRY(Point, 4326),
                to_node GEOMETRY(Point, 4326),
                length DOUBLE PRECISION,
                freespeed DOUBLE PRECISION,
                scale_01_agents TEXT[],
                scale_02_agents TEXT[],
                scale_03_agents TEXT[],
                scale_04_agents TEXT[],
                scale_05_agents TEXT[],
                scale_06_agents TEXT[],
                scale_07_agents TEXT[],
                scale_08_agents TEXT[],
                scale_09_agents TEXT[],
                scale_10_agents TEXT[]
            );
            "#,
        )
        .execute(&self.pool)
        .await
        .context("Failed to create tables")?;

        Ok(())
    }

    pub async fn insert_agent(
        &self,
        id: &str,
        start_coord: &str,
        percentages: &[i32],
        agent_xml: &str,
    ) -> Result<()> {
        sqlx::query(
            r#"
            INSERT INTO agents (id, start_coordinates, percentages, agent_xml)
            VALUES ($1, ST_GeomFromText($2, 4326), $3, $4)
            ON CONFLICT (id) DO UPDATE
            SET start_coordinates = EXCLUDED.start_coordinates,
                percentages = EXCLUDED.percentages,
                agent_xml = EXCLUDED.agent_xml
            "#,
        )
        .bind(id)
        .bind(start_coord)
        .bind(percentages)
        .bind(agent_xml)
        .execute(&self.pool)
        .await
        .context("Failed to insert agent")?;

        Ok(())
    }

    pub async fn insert_link_coordinates(
        &self,
        link_id: &str,
        from_node: &str,
        to_node: &str,
        length: Option<f64>,
        freespeed: Option<f64>,
        scale_agents: &[Vec<String>],
    ) -> Result<()> {
        sqlx::query(
            r#"
            INSERT INTO link_coordinates (
                link_id, from_node, to_node, length, freespeed,
                scale_01_agents, scale_02_agents, scale_03_agents,
                scale_04_agents, scale_05_agents, scale_06_agents,
                scale_07_agents, scale_08_agents, scale_09_agents,
                scale_10_agents
            )
            VALUES (
                $1, ST_GeomFromText($2, 4326), ST_GeomFromText($3, 4326), $4, $5,
                $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
            )
            ON CONFLICT (link_id) DO UPDATE
            SET from_node = EXCLUDED.from_node,
                to_node = EXCLUDED.to_node,
                length = EXCLUDED.length,
                freespeed = EXCLUDED.freespeed,
                scale_01_agents = EXCLUDED.scale_01_agents,
                scale_02_agents = EXCLUDED.scale_02_agents,
                scale_03_agents = EXCLUDED.scale_03_agents,
                scale_04_agents = EXCLUDED.scale_04_agents,
                scale_05_agents = EXCLUDED.scale_05_agents,
                scale_06_agents = EXCLUDED.scale_06_agents,
                scale_07_agents = EXCLUDED.scale_07_agents,
                scale_08_agents = EXCLUDED.scale_08_agents,
                scale_09_agents = EXCLUDED.scale_09_agents,
                scale_10_agents = EXCLUDED.scale_10_agents
            "#,
        )
        .bind(link_id)
        .bind(from_node)
        .bind(to_node)
        .bind(length)
        .bind(freespeed)
        .bind(&scale_agents[0])
        .bind(&scale_agents[1])
        .bind(&scale_agents[2])
        .bind(&scale_agents[3])
        .bind(&scale_agents[4])
        .bind(&scale_agents[5])
        .bind(&scale_agents[6])
        .bind(&scale_agents[7])
        .bind(&scale_agents[8])
        .bind(&scale_agents[9])
        .execute(&self.pool)
        .await
        .context("Failed to insert link coordinates")?;

        Ok(())
    }
}
