package population

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func ConnectToDatabase() (*sql.DB, error) {
	host := getEnvWithDefault("DB_HOST", "localhost")
	port := getEnvWithDefault("DB_PORT", "5432")
	user := getEnvWithDefault("DB_USER", "postgres")
	password := getEnvWithDefault("DB_PASSWORD", "postgres")
	dbname := getEnvWithDefault("DB_NAME", "simulation")

	pgConnStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// Helper function to get environment variable with default value
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
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
