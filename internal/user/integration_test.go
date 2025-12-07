package user

import (
	"context"
	"database/sql"
	"ez2boot/internal/config"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/db"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// Build environment
func newTestServer(t *testing.T) (*Handler, *sql.DB) {
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

	cfg := &config.Config{}

	// User repo constructor
	userRepo := NewRepository(baseRepo, logger)

	// User service constructor
	service := NewService(userRepo, cfg, logger)

	// User handler constructor
	handler := NewHandler(service, logger)

	return handler, testDB
}

func TestCreateUser_Success(t *testing.T) {
	h, db := newTestServer(t)

	// Create admin user which is required for user creation
	query := `INSERT INTO users (id, email, password_hash, is_active, is_admin, api_enabled, ui_enabled, identity_provider)
    		VALUES (1, 'admin@example.com', 'x', 1, 1, 1, 1, 'local')`

	if _, err := db.Exec(query); err != nil {
		t.Fatalf("failed to insert admin user: %v", err)
	}

	body := `{
        "email": "test@example.com",
        "password": "strongpassword123",
        "is_active": true,
        "is_admin": false,
        "api_enabled": true,
        "ui_enabled": true
    }`

	req := httptest.NewRequest("POST", "/user/new", strings.NewReader(body))

	ctx := context.WithValue(req.Context(), ctxutil.UserIDKey, int64(1))

	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	h.CreateUser().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify DB row exists
	var email string
	row := db.QueryRow("SELECT email FROM users WHERE email = ?", "test@example.com")
	if err := row.Scan(&email); err != nil {
		t.Fatalf("user not inserted: %v", err)
	}
	if email != "test@example.com" {
		t.Fatalf("wrong email inserted: %s", email)
	}
}
