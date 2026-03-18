package sqliteutil

import (
	"database/sql"
	"fmt"
)

// DefaultPragmas is the recommended SQLite connection string suffix for
// WAL mode, 5-second busy timeout, and foreign key enforcement.
const DefaultPragmas = "?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on"

// Migration represents a single numbered schema migration.
type Migration struct {
	ID int
	Up func(*sql.DB) error
}

// RunMigrations applies all unapplied migrations in slice order.
// It creates the schema_migrations tracking table if it doesn't exist.
// Callers must provide migrations sorted by ID.
func RunMigrations(conn *sql.DB, migrations []Migration) error {
	if _, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id         INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT (datetime('now'))
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, m := range migrations {
		var n int
		_ = conn.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE id = ?`, m.ID).Scan(&n)
		if n > 0 {
			continue
		}
		if err := m.Up(conn); err != nil {
			return fmt.Errorf("migration %d: %w", m.ID, err)
		}
		if _, err := conn.Exec(`INSERT INTO schema_migrations (id) VALUES (?)`, m.ID); err != nil {
			return fmt.Errorf("record migration %d: %w", m.ID, err)
		}
	}
	return nil
}
