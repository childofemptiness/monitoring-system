package postgres

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultTestDatabaseAdminURL = "postgres://monitor:monitor@localhost:5432/postgres?sslmode=disable"

func setupTestDatabase(t *testing.T) *pgxpool.Pool {
	t.Helper()

	adminURL := os.Getenv("TEST_DATABASE_ADMIN_URL")
	if adminURL == "" {
		adminURL = defaultTestDatabaseAdminURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	adminPool, err := NewPool(ctx, adminURL)
	if err != nil {
		t.Skipf("skip integration test: cannot connect to admin postgres %q: %v", adminURL, err)
	}

	dbName := testDatabaseName()
	if _, err := adminPool.Exec(ctx, "CREATE DATABASE "+dbName); err != nil {
		adminPool.Close()
		t.Fatalf("create test database %s: %v", dbName, err)
	}

	testDatabaseURL, err := databaseURLWithName(adminURL, dbName)
	if err != nil {
		terminateAndDropDatabase(t, adminPool, dbName)
		adminPool.Close()
		t.Fatalf("build test database url: %v", err)
	}

	testPool, err := NewPool(ctx, testDatabaseURL)
	if err != nil {
		terminateAndDropDatabase(t, adminPool, dbName)
		adminPool.Close()
		t.Fatalf("connect test database: %v", err)
	}

	applyMigrations(t, testPool)

	t.Cleanup(func() {
		testPool.Close()
		terminateAndDropDatabase(t, adminPool, dbName)
		adminPool.Close()
	})

	return testPool
}

func applyMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	migrationsDir := migrationsDir(t)
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		path := filepath.Join(migrationsDir, entry.Name())
		query, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read migration %s: %v", entry.Name(), err)
		}

		if _, err := pool.Exec(ctx, string(query)); err != nil {
			t.Fatalf("apply migration %s: %v", entry.Name(), err)
		}
	}
}

func terminateAndDropDatabase(t *testing.T, adminPool *pgxpool.Pool, dbName string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := adminPool.Exec(ctx, `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid()
	`, dbName); err != nil {
		t.Fatalf("terminate connections for %s: %v", dbName, err)
	}

	if _, err := adminPool.Exec(ctx, "DROP DATABASE IF EXISTS "+dbName); err != nil {
		t.Fatalf("drop test database %s: %v", dbName, err)
	}
}

func databaseURLWithName(rawURL, dbName string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse database url: %w", err)
	}

	parsed.Path = "/" + dbName
	return parsed.String(), nil
}

func migrationsDir(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current file path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations"))
}

func testDatabaseName() string {
	return "url_monitor_test_" + strings.ReplaceAll(time.Now().UTC().Format("20060102150405.000000000"), ".", "_")
}
