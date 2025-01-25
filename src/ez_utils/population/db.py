import os
import json
from pathlib import Path
import psycopg2
from psycopg2.extras import execute_values
from typing import Dict, List, Optional

def get_db_connection():
    required_vars = ["POSTGRES_DB", "POSTGRES_USER", "POSTGRES_PASSWORD"]
    missing_vars = [var for var in required_vars if not os.getenv(var)]
    
    if missing_vars:
        raise ValueError(
            f"Missing required environment variables: {', '.join(missing_vars)}. "
            "Please check your .env file."
        )
    
    return psycopg2.connect(
        dbname=os.getenv("POSTGRES_DB"),
        user=os.getenv("POSTGRES_USER"),
        password=os.getenv("POSTGRES_PASSWORD"),
        host=os.getenv("POSTGRES_HOST", "localhost"),
        port=os.getenv("POSTGRES_PORT", "5432")
    )

def ensure_tables_exist():
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Create agents table
            cur.execute("""
                CREATE TABLE IF NOT EXISTS agents (
                    agent_id TEXT PRIMARY KEY,
                    start_coord GEOMETRY(Point, 4326),
                    percentages SMALLINT[] NOT NULL,
                    agent_xml TEXT
                );
            """)
            
            # Create spatial index for agents
            cur.execute("""
                CREATE INDEX IF NOT EXISTS idx_agents_spatial 
                ON agents 
                USING gist (start_coord);
            """)
            
            # Create GIN index for percentages array
            cur.execute("""
                CREATE INDEX IF NOT EXISTS idx_agents_percentages 
                ON agents 
                USING GIN (percentages);
            """)
            
            # Create link_coordinates table
            cur.execute("""
                CREATE TABLE IF NOT EXISTS link_coordinates (
                    link_id TEXT PRIMARY KEY,
                    from_node GEOMETRY(Point, 4326) NOT NULL,
                    to_node GEOMETRY(Point, 4326) NOT NULL,
                    length DOUBLE PRECISION,
                    freespeed DOUBLE PRECISION,
                    pass_01pct TEXT[],
                    pass_02pct TEXT[],
                    pass_03pct TEXT[],
                    pass_04pct TEXT[],
                    pass_05pct TEXT[],
                    pass_06pct TEXT[],
                    pass_07pct TEXT[],
                    pass_08pct TEXT[],
                    pass_09pct TEXT[],
                    pass_10pct TEXT[]
                );
            """)
            
            # Create spatial index for link_coordinates
            cur.execute("""
                CREATE INDEX IF NOT EXISTS idx_link_coordinates_spatial 
                ON link_coordinates 
                USING gist (from_node, to_node);
            """)
            
            conn.commit()

def insert_agents(agent_data: List[tuple]):
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            execute_values(
                cur,
                """
                INSERT INTO agents (agent_id, start_coord, percentages, agent_xml)
                VALUES %s
                ON CONFLICT (agent_id) DO UPDATE SET
                    start_coord = EXCLUDED.start_coord,
                    percentages = EXCLUDED.percentages,
                    agent_xml = EXCLUDED.agent_xml
                """,
                agent_data
            )
            conn.commit()

def insert_link_coordinates(link_data: List[tuple]):
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            execute_values(
                cur,
                """
                INSERT INTO link_coordinates (
                    link_id, from_node, to_node, length, freespeed,
                    pass_01pct, pass_02pct, pass_03pct, pass_04pct, pass_05pct,
                    pass_06pct, pass_07pct, pass_08pct, pass_09pct, pass_10pct
                )
                VALUES %s
                ON CONFLICT (link_id) DO UPDATE SET
                    from_node = EXCLUDED.from_node,
                    to_node = EXCLUDED.to_node,
                    length = EXCLUDED.length,
                    freespeed = EXCLUDED.freespeed,
                    pass_01pct = EXCLUDED.pass_01pct,
                    pass_02pct = EXCLUDED.pass_02pct,
                    pass_03pct = EXCLUDED.pass_03pct,
                    pass_04pct = EXCLUDED.pass_04pct,
                    pass_05pct = EXCLUDED.pass_05pct,
                    pass_06pct = EXCLUDED.pass_06pct,
                    pass_07pct = EXCLUDED.pass_07pct,
                    pass_08pct = EXCLUDED.pass_08pct,
                    pass_09pct = EXCLUDED.pass_09pct,
                    pass_10pct = EXCLUDED.pass_10pct
                """,
                link_data
            )
            conn.commit()
