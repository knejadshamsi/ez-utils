package gui

// Table creation queries
const (
	createProcessesTableQuery = `
		CREATE TABLE IF NOT EXISTS processes (
			process_id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_path TEXT NOT NULL,
			status TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)`

	createTelemetryTableQuery = `
		CREATE TABLE IF NOT EXISTS process_telemetry (
			process_id INTEGER PRIMARY KEY,
			total_file_size INTEGER NOT NULL,
			bytes_read INTEGER DEFAULT 0,
			persons_extracted INTEGER DEFAULT 0,
			error_count INTEGER DEFAULT 0,
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (process_id) REFERENCES processes(process_id) ON DELETE CASCADE
		)`

	createPopulationTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			coords TEXT,
			raw_xml TEXT
		)`

	dropTableQuery = "DROP TABLE IF EXISTS %s"
)
