package sqlite

import (
	"database/sql"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	migratesqlite "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// The .sql files are compiled into the binary, so the program carries its own schema.
//
//go:embed migrations/*.sql
var migrationFiles embed.FS

// runMigrations applies every migration that has not been applied yet.
// The current version is stored in a table called schema_migrations.
func runMigrations(db *sql.DB) error {

	source, err := iofs.New(migrationFiles, "migrations")

	if err != nil {
		return err
	}

	driver, err := migratesqlite.WithInstance(db, &migratesqlite.Config{})

	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", source, "sqlite3", driver)

	if err != nil {
		return err
	}

	err = m.Up()

	// ErrNoChange just means the database is already up to date.
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
