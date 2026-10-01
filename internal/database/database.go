package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	_ "modernc.org/sqlite"
)

type Database struct {
	DB *sql.DB
}

func Connect(ctx context.Context, path string) (*Database, error) {
	// Enable WAL mode and foreign keys for better performance and safety
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)", path)
	
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Connected to database", "path", path)
	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

// Migrate would run schema migrations.
// In Phase 1 we will just create the tables if they don't exist.
func (d *Database) Migrate(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS tokens (
		id TEXT PRIMARY KEY,
		description TEXT,
		token_hash TEXT NOT NULL UNIQUE,
		permissions TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		revoked BOOLEAN DEFAULT FALSE
	);

	CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		token_id TEXT,
		action TEXT NOT NULL,
		target TEXT,
		parameters TEXT,
		result TEXT,
		error_message TEXT,
		client_ip TEXT,
		FOREIGN KEY (token_id) REFERENCES tokens(id)
	);
	`
	
	_, err := d.DB.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	slog.Info("Database migrations applied successfully")
	return nil
}
