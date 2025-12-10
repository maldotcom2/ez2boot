package testutil

import (
	"database/sql"
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"io"
	"log/slog"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

type TestEnv struct {
	DB     *sql.DB
	Logger *slog.Logger
	Base   *db.Repository
	Cfg    *config.Config
}

// Build test environment - in memory only
func NewTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	// Create in-memory sqlite
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// Enable foreign keys
	_, err = testDB.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		t.Fatal(err)
	}

	// Enable WAL
	_, err = testDB.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// DB base Constructor
	baseRepo := db.NewRepository(testDB, logger)

	// Setup schema
	if err := baseRepo.SetupDB(); err != nil {
		t.Fatalf("failed to setup DB: %v", err)
	}

	env := &TestEnv{
		DB:     testDB,
		Logger: logger,
		Base:   baseRepo,
		Cfg:    &config.Config{},
	}

	return env
}
