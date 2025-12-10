package user

import (
	"context"
	"database/sql"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/testutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestCreateUser_Success(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Domain constructors
	userRepo := NewRepository(env.Base, env.Logger)
	userSvc := NewService(userRepo, env.Cfg, env.Logger)
	handler := NewHandler(userSvc, env.Logger)

	// Create admin user which is required for user creation
	query := `INSERT INTO users (id, email, password_hash, is_active, is_admin, api_enabled, ui_enabled, identity_provider)
    		VALUES (1, 'admin@example.com', 'x', 1, 1, 1, 1, 'local')`

	if _, err := env.DB.Exec(query); err != nil {
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

	handler.CreateUser().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify DB row exists
	var email string
	row := env.DB.QueryRow("SELECT email FROM users WHERE email = ?", "test@example.com")
	if err := row.Scan(&email); err != nil {
		t.Fatalf("user not inserted: %v", err)
	}
	if email != "test@example.com" {
		t.Fatalf("wrong email inserted: %s", email)
	}
}

func TestCreateUser_NotAdmin_ReturnsForbidden(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Domain constructors
	userRepo := NewRepository(env.Base, env.Logger)
	userSvc := NewService(userRepo, env.Cfg, env.Logger)
	handler := NewHandler(userSvc, env.Logger)

	// Create non-admin user
	query := `INSERT INTO users (id, email, password_hash, is_active, is_admin, api_enabled, ui_enabled, identity_provider)
    		VALUES (1, 'admin@example.com', 'x', 1, 0, 1, 1, 'local')`

	if _, err := env.DB.Exec(query); err != nil {
		t.Fatalf("failed to insert user: %v", err)
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

	handler.CreateUser().ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d, body=%s", w.Code, w.Body.String())
	}

	// Check if user was created anyway
	var email string
	row := env.DB.QueryRow("SELECT email FROM users WHERE email = ?", "test@example.com")
	err := row.Scan(&email)
	if err == nil {
		t.Fatalf("user was created by non-admin: %s", email)
	}
	if err != sql.ErrNoRows {
		t.Fatalf("unexpected DB error: %v", err)
	}
}
