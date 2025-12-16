package testutil

import (
	"database/sql"
	"ez2boot/internal/app"
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"io"
	"log/slog"
	"net/http"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

type TestEnv struct {
	DB     *sql.DB
	Logger *slog.Logger
	Base   *db.Repository
	Cfg    *config.Config
	Router http.Handler
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

	cfg := &config.Config{
		RateLimit: 30,
	}

	router, _, _, err := app.NewApp("dev", "unknown", cfg, baseRepo, logger)
	if err != nil {
		t.Fatalf("failed to initialize app: %v", err)
	}

	return &TestEnv{
		DB:     testDB,
		Logger: logger,
		Base:   baseRepo,
		Cfg:    cfg,
		Router: router,
	}
}

func InsertUser(t *testing.T, db *sql.DB, email string, passwordHash string, isActive bool, isAdmin bool, apiEnabled bool, uiEnabled bool) {
	t.Helper()

	_, err := db.Exec(`
        INSERT INTO users (email, password_hash, is_active, is_admin, api_enabled, ui_enabled, identity_provider)
        VALUES ($1, $2, $3, $4, $5, $6, 'local')
    `, email, passwordHash, isActive, isAdmin, apiEnabled, uiEnabled)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
}
