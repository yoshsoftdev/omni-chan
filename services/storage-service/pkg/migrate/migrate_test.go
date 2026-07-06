package migrate

import (
	"strings"
	"testing"
)

func TestWithMigrationsTable(t *testing.T) {
	got := withMigrationsTable("postgres://user:pass@localhost:5432/db?sslmode=disable")
	if !strings.Contains(got, "x-migrations-table=storage_schema_migrations") {
		t.Fatalf("expected storage migrations table in URL, got %q", got)
	}
}
