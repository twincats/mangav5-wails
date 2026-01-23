package db

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func Migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY
		)
	`)
	if err != nil {
		return err
	}

	rows, _ := db.Query(`SELECT version FROM schema_migrations`)
	applied := map[string]bool{}
	for rows.Next() {
		var v string
		rows.Scan(&v)
		applied[v] = true
	}
	rows.Close()

	files, _ := migrationFS.ReadDir("migrations")

	var names []string
	for _, f := range files {
		names = append(names, f.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		if applied[name] {
			continue
		}

		sqlBytes, _ := migrationFS.ReadFile("migrations/" + name)

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s failed: %w", name, err)
		}

		if _, err := tx.Exec(
			`INSERT INTO schema_migrations(version) VALUES (?)`,
			name,
		); err != nil {
			tx.Rollback()
			return err
		}

		tx.Commit()
	}

	return nil
}
