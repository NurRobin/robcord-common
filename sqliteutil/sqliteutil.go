package sqliteutil

import (
	"context"
	"database/sql"
	"fmt"
)

// CheckWritable verifies the database is writable by performing a test write cycle.
// Uses INSERT OR REPLACE for idempotency (handles interrupted previous checks).
func CheckWritable(ctx context.Context, conn *sql.DB) error {
	_, err := conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS _health_check (id INTEGER PRIMARY KEY)`)
	if err != nil {
		return fmt.Errorf("database is not writable (check file ownership/permissions): %w", err)
	}
	_, err = conn.ExecContext(ctx, `INSERT OR REPLACE INTO _health_check (id) VALUES (1)`)
	if err != nil {
		return fmt.Errorf("database is not writable (INSERT failed): %w", err)
	}
	_, err = conn.ExecContext(ctx, `DELETE FROM _health_check WHERE id = 1`)
	if err != nil {
		return fmt.Errorf("database is not writable (DELETE failed): %w", err)
	}
	return nil
}

// ColumnExists checks whether a column exists in a SQLite table.
func ColumnExists(conn *sql.DB, table, column string) bool {
	rows, err := conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var dfltValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &dfltValue, &pk); err != nil {
			return false
		}
		if name == column {
			return true
		}
	}
	return false
}

// AlterIfMissing checks PRAGMA table_info for the given column and only
// executes the DDL (ALTER TABLE) if the column does not exist yet.
func AlterIfMissing(conn *sql.DB, table, column, ddl string) error {
	if ColumnExists(conn, table, column) {
		return nil
	}
	_, err := conn.Exec(ddl)
	return err
}
