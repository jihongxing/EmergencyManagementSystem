package database

import (
	"context"
	"emergency-management/backend/internal/platform"
	"net/http/httptest"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/pressly/goose/v3"
)

func TestInvalidConfiguration(t *testing.T) {
	for _, url := range []string{"", "postgres://%zz:secret@"} {
		if db, err := Open(url); err == nil {
			db.Close()
			t.Fatal("invalid configuration accepted")
		}
	}
}

func TestUnavailable(t *testing.T) {
	db, err := Open("postgres://nobody:private@127.0.0.1:1/missing?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if Ready(db)(ctx) == nil {
		t.Fatal("unreachable database ready")
	}
}

func TestIntegration(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL required for dedicated integration database")
	}
	db, err := Open(url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// The caller provisions an empty dedicated database; never reset shared data.
	var exists bool
	if err := db.QueryRowContext(ctx, "SELECT to_regclass('public.goose_db_version') IS NOT NULL").Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("integration database must be empty")
	}
	if Ready(db)(ctx) == nil {
		t.Fatal("unmigrated database ready")
	}
	assertHTTP := func(expected int) {
		t.Helper()
		rec := httptest.NewRecorder()
		platform.NewHandler(Ready(db)).ServeHTTP(rec, httptest.NewRequest("GET", "/health/ready", nil))
		if rec.Code != expected {
			t.Fatalf("readiness HTTP: got %d want %d", rec.Code, expected)
		}
	}
	assertHTTP(503)
	for range 2 {
		if err := Migrate(ctx, db); err != nil {
			t.Fatal(err)
		}
	}
	if err := Ready(db)(ctx); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{
		"ems.organizations",
		"ems.members",
		"ems.organization_materials",
		"ems.identity_audit_events",
		"ems.sessions",
		"ems.password_resets",
		"ems.member_audit_events",
	} {
		if err := db.QueryRowContext(ctx, "SELECT to_regclass($1) IS NOT NULL", table).Scan(&exists); err != nil || !exists {
			t.Fatalf("identity bootstrap table missing: %s (%v)", table, err)
		}
	}
	assertHTTP(200)
	files := fstest.MapFS{
		"00002_failure.sql": &fstest.MapFile{Data: []byte("-- +goose Up\nCREATE TABLE ems.must_rollback (id int);\nSELECT 1/0;\n")},
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, files)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = provider.Up(ctx); err == nil {
		t.Fatal("broken migration succeeded")
	}
	if err := db.QueryRowContext(ctx, "SELECT to_regclass('ems.must_rollback') IS NOT NULL").Scan(&exists); err != nil || exists {
		t.Fatalf("failed migration left table: %v %v", exists, err)
	}
	if err := Ready(db)(ctx); err != nil {
		t.Fatal("failed migration changed baseline")
	}
}

func TestPersistedReadiness(t *testing.T) {
	url := os.Getenv("TEST_PERSISTED_DATABASE_URL")
	if url == "" {
		t.Skip("dedicated migrated database required")
	}
	db, err := Open(url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rec := httptest.NewRecorder()
	platform.NewHandler(Ready(db)).ServeHTTP(rec, httptest.NewRequest("GET", "/health/ready", nil))
	if rec.Code != 200 {
		t.Fatalf("readiness did not recover: %d", rec.Code)
	}
}
