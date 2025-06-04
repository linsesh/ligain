package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testDB struct {
	db        *sql.DB
	container testcontainers.Container
}

func setupTestDB(t *testing.T) *testDB {
	log.Println("Starting test database setup...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start PostgreSQL container
	log.Println("Starting PostgreSQL container...")
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "ligain_test",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort(nat.Port("5432/tcp")),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	log.Println("PostgreSQL container started successfully")

	// Get container port
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	// Connect to test database
	dbURL := fmt.Sprintf("postgres://postgres:postgres@localhost:%s/ligain_test?sslmode=disable", port.Port())
	log.Printf("Connecting to database at %s", dbURL)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Set connection timeout
	db.SetConnMaxLifetime(30 * time.Second)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	// Wait for database to be ready with timeout
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()

	pingDone := make(chan error, 1)
	go func() {
		pingDone <- db.Ping()
	}()

	select {
	case err := <-pingDone:
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
	case <-pingCtx.Done():
		t.Fatalf("Timeout waiting for database connection")
	}
	log.Println("Database connection established")

	// Create test database instance
	testDB := &testDB{
		db:        db,
		container: container,
	}

	// Clean up any existing data
	testDB.cleanup(t)

	return testDB
}

func (db *testDB) cleanup(t *testing.T) {
	log.Println("Starting database cleanup...")
	// Drop all tables
	_, err := db.db.Exec(`
		DROP TABLE IF EXISTS score CASCADE;
		DROP TABLE IF EXISTS bet CASCADE;
		DROP TABLE IF EXISTS match CASCADE;
		DROP TABLE IF EXISTS player CASCADE;
		DROP TABLE IF EXISTS game CASCADE;
		DROP TABLE IF EXISTS schema_migrations CASCADE;
	`)
	if err != nil {
		t.Fatalf("Failed to clean up test database: %v", err)
	}
	log.Println("Tables dropped successfully")

	// Get database URL for migration
	port, err := db.container.MappedPort(context.Background(), "5432")
	if err != nil {
		t.Fatalf("Failed to get container port for migration: %v", err)
	}
	dbURL := fmt.Sprintf("postgres://postgres:postgres@localhost:%s/ligain_test?sslmode=disable", port.Port())

	// Create separate connection for migration
	migrationDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to create migration database connection: %v", err)
	}
	defer migrationDB.Close()

	// Initialize golang-migrate with separate connection
	log.Println("Initializing golang-migrate...")
	driver, err := postgres.WithInstance(migrationDB, &postgres.Config{})
	if err != nil {
		t.Fatalf("Failed to create postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres",
		driver,
	)
	if err != nil {
		t.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			t.Errorf("Failed to close migration source: %v", sourceErr)
		}
		if dbErr != nil {
			t.Errorf("Failed to close migration database: %v", dbErr)
		}
	}()

	// Run migrations with timeout
	log.Println("Running migrations...")
	migrationCtx, migrationCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer migrationCancel()

	migrationDone := make(chan error, 1)
	go func() {
		migrationDone <- m.Up()
	}()

	select {
	case err := <-migrationDone:
		if err != nil && err != migrate.ErrNoChange {
			t.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Migrations completed successfully")
	case <-migrationCtx.Done():
		t.Fatalf("Timeout waiting for migrations to complete")
	}
}

func (db *testDB) withTransaction(t *testing.T, testFunc func(*sql.Tx)) {
	tx, err := db.db.Begin()
	require.NoError(t, err)

	defer func() {
		if r := recover(); r != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
				t.Logf("Failed to rollback transaction after panic: %v", rollbackErr)
			}
			panic(r)
		}
	}()

	testFunc(tx)

	if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
		t.Errorf("Failed to rollback transaction: %v", rollbackErr)
	}
}

func (db *testDB) Close() error {
	if err := db.db.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %v", err)
	}
	if err := db.container.Terminate(context.Background()); err != nil {
		return fmt.Errorf("failed to terminate container: %v", err)
	}
	return nil
}

// runTestWithTimeout runs a test function with a timeout.
// This is a safety net for catching unexpected infinite loops or deadlocks.
func runTestWithTimeout(t *testing.T, testFunc func(t *testing.T), timeout time.Duration) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		testFunc(t)
	}()

	select {
	case <-time.After(timeout):
		t.Fatal("Test timed out. This likely indicates a deadlock or infinite loop.")
	case <-done:
		// Test completed successfully
	}
}
