package migrate

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func Up(db *sql.DB, dir string) error {
	goose.SetTableName("schema_migrations")

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func DownTo(db *sql.DB, dir string, version int64) error {
	goose.SetTableName("schema_migrations")

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.DownTo(db, dir, version); err != nil {
		return fmt.Errorf("rollback migrations to %d: %w", version, err)
	}

	return nil
}
