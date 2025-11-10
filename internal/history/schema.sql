-- Query history schema
CREATE TABLE IF NOT EXISTS query_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    connection_name TEXT,
    database_name TEXT,
    query TEXT NOT NULL,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    duration_ms INTEGER,
    rows_affected INTEGER,
    success BOOLEAN NOT NULL,
    error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_executed_at ON query_history(executed_at DESC);
CREATE INDEX IF NOT EXISTS idx_connection ON query_history(connection_name);
CREATE INDEX IF NOT EXISTS idx_success ON query_history(success);
