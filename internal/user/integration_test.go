package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"ez2boot/internal/testutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestGetUsers_Success(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Domain constructors
	userRepo := NewRepository(env.Base, env.Logger)
	userSvc := NewService(userRepo, env.Cfg, env.Logger)
	handler := NewHandler(userSvc, env.Logger)

	// Create users
	testutil.InsertUser(t, env.DB, "example@example.com", false)
	testutil.InsertUser(t, env.DB, "example2@example.com", false)

	// Call endpoint
	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	handler.GetUsers().ServeHTTP(w, req)

	// Code check
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var got shared.ApiResponse[[]User]

	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Expected API response
	want := shared.ApiResponse[[]User]{
		Success: true,
		Data: []User{
			{
				UserID:     1,
				Email:      "example@example.com",
				IsActive:   true,
				IsAdmin:    false,
				APIEnabled: true,
				UIEnabled:  true,
				LastLogin:  nil,
			},
			{
				UserID:     2,
				Email:      "example2@example.com",
				IsActive:   true,
				IsAdmin:    false,
				APIEnabled: true,
				UIEnabled:  true,
				LastLogin:  nil,
			},
		},
		Error: "",
	}

	// Compare response body
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("body mismatch\n got:  %#v\n want: %#v", got, want)
	}

}

func TestGetUserAuthorisation_Success(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Domain constructors
	userRepo := NewRepository(env.Base, env.Logger)
	userSvc := NewService(userRepo, env.Cfg, env.Logger)
	handler := NewHandler(userSvc, env.Logger)

	// Create target user
	testutil.InsertUser(t, env.DB, "example@example.com", false)

	// Call endpoint with required userID in context
	req := httptest.NewRequest("GET", "/user/auth", nil)
	ctx := context.WithValue(req.Context(), ctxutil.UserIDKey, int64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.GetUserAuthorisation().ServeHTTP(w, req)

	// Code check
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var got shared.ApiResponse[UserAuthRequest]

	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Expected API response
	want := shared.ApiResponse[UserAuthRequest]{
		Success: true,
		Data: UserAuthRequest{
			UserID:     1,
			Email:      "example@example.com",
			IsActive:   true,
			IsAdmin:    false,
			APIEnabled: true,
			UIEnabled:  true,
		},
		Error: "",
	}

	// Compare response body
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("body mismatch\n got:  %#v\n want: %#v", got, want)
	}

}

func TestCreateUser_Success(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Domain constructors
	userRepo := NewRepository(env.Base, env.Logger)
	userSvc := NewService(userRepo, env.Cfg, env.Logger)
	handler := NewHandler(userSvc, env.Logger)

	// Create admin user which is required for user creation
	testutil.InsertUser(t, env.DB, "admin@example.com", true)

	// Request body
	body := `{
        "email": "test@example.com",
        "password": "strongpassword123",
        "is_active": true,
        "is_admin": false,
        "api_enabled": true,
        "ui_enabled": true
    }`

	// Call endpoint with required userID in context
	req := httptest.NewRequest("POST", "/user/new", strings.NewReader(body))
	ctx := context.WithValue(req.Context(), ctxutil.UserIDKey, int64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.CreateUser().ServeHTTP(w, req)

	// Code check
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
	testutil.InsertUser(t, env.DB, "example@example.com", false)

	// Request body
	body := `{
        "email": "test@example.com",
        "password": "strongpassword123",
        "is_active": true,
        "is_admin": false,
        "api_enabled": true,
        "ui_enabled": true
    }`

	// Call endpoint with required userID in context
	req := httptest.NewRequest("POST", "/user/new", strings.NewReader(body))
	ctx := context.WithValue(req.Context(), ctxutil.UserIDKey, int64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.CreateUser().ServeHTTP(w, req)

	// Code check
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
