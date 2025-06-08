package population

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"ez-utils/src/config"
	"fmt"

	_ "github.com/lib/pq"
)

func ConnectToDatabase() (*sql.DB, error) {
	// Get configuration
	cfg := config.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("configuration not loaded - please ensure config.yaml is properly configured")
	}

	// Validate database configuration
	if cfg.Database.Host == "" {
		return nil, fmt.Errorf("database host not configured in config.yaml")
	}
	if cfg.Database.User == "" {
		return nil, fmt.Errorf("database user not configured in config.yaml")
	}
	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("database password not configured in config.yaml")
	}
	if cfg.Database.Name == "" {
		return nil, fmt.Errorf("database name not configured in config.yaml")
	}

	// Use config values
	pgConnStr := cfg.Database.GetConnectionString()

	db, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure pool settings from config
	db.SetMaxOpenConns(cfg.Database.MaxConnections)
	db.SetMaxIdleConns(cfg.Database.MaxConnections)

	timeout, err := cfg.Database.GetConnectionTimeout()
	if err != nil {
		return nil, fmt.Errorf("invalid connection timeout in config: %w", err)
	}
	db.SetConnMaxLifetime(timeout)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func CreateAgentsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS agents (
			agent_id TEXT PRIMARY KEY,
			start_coord GEOMETRY(Point, 4326),
			scale_1pct BOOLEAN NOT NULL DEFAULT FALSE,
			scale_2pct BOOLEAN NOT NULL DEFAULT FALSE,
			scale_3pct BOOLEAN NOT NULL DEFAULT FALSE,
			scale_4pct BOOLEAN NOT NULL DEFAULT FALSE,
			scale_5pct BOOLEAN NOT NULL DEFAULT FALSE,
			scale_6pct BOOLEAN NOT NULL DEFAULT FALSE,
			scale_7pct BOOLEAN NOT NULL DEFAULT FALSE,
			scale_8pct BOOLEAN NOT NULL DEFAULT FALSE,
			scale_9pct BOOLEAN NOT NULL DEFAULT FALSE,
			scale_10pct BOOLEAN NOT NULL DEFAULT FALSE,
			agent_xml BYTEA
		)
	`)
	return err
}

func InsertAgentBatch(db *sql.DB, agents []Agent) error {
	stmt, err := db.Prepare(`
		INSERT INTO agents 
		(agent_id, start_coord, scale_1pct, scale_2pct, scale_3pct, scale_4pct, 
		 scale_5pct, scale_6pct, scale_7pct, scale_8pct, scale_9pct, scale_10pct, agent_xml) 
		VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), false, false, false, false, 
				false, false, false, false, false, false, $4)
		ON CONFLICT (agent_id) DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Set statement timeout to prevent hanging operations
	_, err = db.Exec("SET statement_timeout = '30s'")
	if err != nil {
		return err
	}

	batchSize := 1000
	for i := 0; i < len(agents); i += batchSize {
		tx, err := db.Begin()
		if err != nil {
			return err
		}

		end := min(i+batchSize, len(agents))
		for j := i; j < end; j++ {
			var buffer bytes.Buffer
			gzipWriter := gzip.NewWriter(&buffer)
			// Use empty string if XML is not provided
			xmlData := agents[j].XML
			if xmlData == "" {
				xmlData = "<person id=\"" + agents[j].ID + "\"/>"
			}

			_, err = gzipWriter.Write([]byte(xmlData))
			if err != nil {
				tx.Rollback()
				return err
			}
			err = gzipWriter.Close()
			if err != nil {
				tx.Rollback()
				return err
			}

			// Execute insert
			_, err = tx.Stmt(stmt).Exec(
				agents[j].ID,
				agents[j].Longitude,
				agents[j].Latitude,
				buffer.Bytes(),
			)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		// Commit transaction
		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
