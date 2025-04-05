package steps

import (
"bytes"
"compress/gzip"
"database/sql"
"fmt"
"os"
"time"

_ "github.com/lib/pq"
)

// ConnectToDatabase establishes a connection to PostgreSQL
func connectToDatabase() (*sql.DB, error) {
// Get connection parameters from environment variables with defaults
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

// Configure pool settings
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

// CreateAgentsTable creates the agents table if it doesn't exist
func createAgentsTable(db *sql.DB) error {
// Create agents table with PostGIS extension
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

// InsertAgentBatch inserts a batch of agents into the database
func insertAgentBatch(db *sql.DB, agents []Agent) error {
// Prepare batch insert statement
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

// Process in batches
batchSize := 1000
for i := 0; i < len(agents); i += batchSize {
// Start transaction
tx, err := db.Begin()
if err != nil {
return err
}

// Process batch
end := min(i+batchSize, len(agents))
for j := i; j < end; j++ {
// Create minimal XML if not provided
var xmlContent string
if agents[j].XML == "" {
xmlContent = fmt.Sprintf("<person id=\"%s\"/>", agents[j].ID)
} else {
xmlContent = agents[j].XML
}

// Compress XML data
var buffer bytes.Buffer
gzipWriter := gzip.NewWriter(&buffer)
if _, err := gzipWriter.Write([]byte(xmlContent)); err != nil {
tx.Rollback()
return err
}
if err := gzipWriter.Close(); err != nil {
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
if err = tx.Commit(); err != nil {
return err
}
}

return nil
}

// Helper function to get the minimum of two integers
func min(a, b int) int {
if a < b {
return a
}
return b
}
