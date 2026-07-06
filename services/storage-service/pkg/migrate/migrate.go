package migrate

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const migrationsTable = "storage_schema_migrations"

func Up(databaseURL, migrationsPath string) error {
	return run(databaseURL, migrationsPath, func(m *migrate.Migrate) error {
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		return nil
	})
}

func Down(databaseURL, migrationsPath string) error {
	return run(databaseURL, migrationsPath, func(m *migrate.Migrate) error {
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		return nil
	})
}

func Force(databaseURL, migrationsPath string, version int) error {
	return run(databaseURL, migrationsPath, func(m *migrate.Migrate) error {
		return m.Force(version)
	})
}

func run(databaseURL, migrationsPath string, fn func(*migrate.Migrate) error) error {
	abs, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}
	if _, err := os.Stat(abs); err != nil {
		return fmt.Errorf("migrations path: %w", err)
	}

	sourceURL := "file://" + filepath.ToSlash(abs)
	m, err := migrate.New(sourceURL, withMigrationsTable(databaseURL))
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	if err := fn(m); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

func withMigrationsTable(databaseURL string) string {
	if strings.Contains(databaseURL, "x-migrations-table=") {
		return databaseURL
	}
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		sep := "?"
		if strings.Contains(databaseURL, "?") {
			sep = "&"
		}
		return databaseURL + sep + "x-migrations-table=" + migrationsTable
	}
	query := parsed.Query()
	query.Set("x-migrations-table", migrationsTable)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
