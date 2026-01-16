package testutil

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"ez2boot/internal/app"
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/user"
	"ez2boot/internal/worker"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type TestEnv struct {
	DB     *sql.DB
	Logger *slog.Logger
	Base   *db.Repository
	Cfg    *config.Config
	Router http.Handler
	Worker *worker.Worker
}

// Build test environment - in memory only
func NewTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	// Create in-memory sqlite
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = testDB.Close()
	})

	// Enable foreign keys
	_, err = testDB.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		t.Fatal(err)
	}

	/* 	// Enable WAL
	   	_, err = testDB.Exec("PRAGMA journal_mode = WAL;")
	   	if err != nil {
	   		t.Fatal(err)
	   	} */

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// DB base Constructor
	baseRepo := db.NewRepository(testDB, logger)

	// Setup schema
	if err := baseRepo.SetupDB(); err != nil {
		t.Fatalf("failed to setup DB: %v", err)
	}

	cfg := &config.Config{
		RateLimit:           30,
		UserSessionDuration: 1 * time.Hour, // Prevent intermittent 401s during test
	}

	router, _, wkr, err := app.NewApp("dev", "unknown", cfg, baseRepo, logger)
	if err != nil {
		t.Fatalf("failed to initialize app: %v", err)
	}

	return &TestEnv{
		DB:     testDB,
		Logger: logger,
		Base:   baseRepo,
		Cfg:    cfg,
		Router: router,
		Worker: wkr,
	}
}

// Insert a dummy user into test database
func InsertUser(t *testing.T, db *sql.DB, email string, passwordHash string, isActive bool, isAdmin bool, apiEnabled bool, uiEnabled bool) {
	t.Helper()

	_, err := db.Exec(`INSERT INTO users (email, password_hash, is_active, is_admin, api_enabled, ui_enabled, identity_provider)
        				VALUES ($1, $2, $3, $4, $5, $6, 'local')`, email, passwordHash, isActive, isAdmin, apiEnabled, uiEnabled)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
}

// Insert a dummy server session into test database
func InsertServer(t *testing.T, db *sql.DB, uniqueID string, name string, state string, serverGroup string, timeAdded int64) {
	t.Helper()

	_, err := db.Exec(`INSERT INTO servers (unique_id, name, state, server_group, time_added)
        				VALUES ($1, $2, $3, $4, $5)`, uniqueID, name, state, serverGroup, timeAdded)
	if err != nil {
		t.Fatalf("failed to insert server: %v", err)
	}
}

// Logs in a UI user and returns cookie
func LoginAndGetCookies(t *testing.T, router http.Handler, email, password string) []*http.Cookie {
	t.Helper()

	payload := user.UserLogin{
		Email:    email,
		Password: password,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/ui/user/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login failed, expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("login did not set a cookie")
	}

	return cookies
}

// Inserts a pending new server session
func InsertServerSession(t *testing.T, db *sql.DB, userID int64, serverGroup string, expiry int64) {
	t.Helper()

	// Set server table for state worker
	if _, err := db.Exec("UPDATE servers SET next_state = $1, time_last_on = $2, last_user_id = $3 WHERE server_group = $4", "on", time.Now().Unix(), userID, serverGroup); err != nil {
		t.Fatal("failed to update server state")
	}

	// Set server session
	if _, err := db.Exec("INSERT INTO server_sessions (user_id, server_group, expiry, warning_notified, on_notified) VALUES ($1, $2, $3, $4, $5)", userID, serverGroup, expiry, 0, 0); err != nil {
		t.Fatal("failed to insert new server session")
	}
}
