package user_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"ez2boot/internal/app"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/router"
	"ez2boot/internal/shared"
	"ez2boot/internal/testutil"
	"ez2boot/internal/user"
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

	// Initialise services, handlers, mw
	mw, _, handlers, _ := app.InitServices("dev", "unknown", env.Cfg, env.Base, env.Logger)

	// Build router
	router := router.BuildRouter(env.Cfg, mw, handlers)

	// Create users
	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	adminHash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, adminHash, true, true, true, true)
	testutil.InsertUser(t, env.DB, "example@example.com", "x", true, false, true, true)
	testutil.InsertUser(t, env.DB, "example2@example.com", "x", true, false, true, true)

	// Login
	loginPayload := user.UserLogin{
		Email:    adminEmail,
		Password: adminPassword,
	}
	loginBody, _ := json.Marshal(loginPayload)
	loginReq := httptest.NewRequest("POST", "/ui/user/login", bytes.NewReader(loginBody))

	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("login failed, expected 200, got %d, body=%s", loginRec.Code, loginRec.Body.String())
	}

	cookies := loginRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("login did not set a cookie")
	}

	// Prepare HTTP request to the real route
	req := httptest.NewRequest("GET", "/ui/users", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	// Record the response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check HTTP status code
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// Decode response
	var got shared.ApiResponse[[]user.User]
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Zero last login value for test comparison
	for i := range got.Data {
		got.Data[i].LastLogin = nil
	}

	// Expected API response
	want := shared.ApiResponse[[]user.User]{
		Success: true,
		Data: []user.User{
			{
				UserID:     1,
				Email:      "admin@example.com",
				IsActive:   true,
				IsAdmin:    true,
				APIEnabled: true,
				UIEnabled:  true,
				LastLogin:  nil,
			},
			{
				UserID:     2,
				Email:      "example@example.com",
				IsActive:   true,
				IsAdmin:    false,
				APIEnabled: true,
				UIEnabled:  true,
				LastLogin:  nil,
			},
			{
				UserID:     3,
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
	userRepo := user.NewRepository(env.Base, env.Logger)
	userSvc := user.NewService(userRepo, env.Cfg, env.Logger)
	handler := user.NewHandler(userSvc, env.Logger)

	// Create target user
	testutil.InsertUser(t, env.DB, "example@example.com", "x", true, false, true, true)

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

	var got shared.ApiResponse[user.UserAuthRequest]

	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Expected API response
	want := shared.ApiResponse[user.UserAuthRequest]{
		Success: true,
		Data: user.UserAuthRequest{
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

func TestUpdateUserAuthorisation_Success(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Domain constructors
	userRepo := user.NewRepository(env.Base, env.Logger)
	userSvc := user.NewService(userRepo, env.Cfg, env.Logger)
	handler := user.NewHandler(userSvc, env.Logger)

	// Create users, caller must be admin
	testutil.InsertUser(t, env.DB, "admin@example.com", "x", true, true, true, true)
	testutil.InsertUser(t, env.DB, "example@example.com", "x", true, false, true, true)

	// Request body
	body := `[{
		"user_id": 2,
        "is_active": true,
        "is_admin": false,
        "api_enabled": false,
        "ui_enabled": true
    }]`

	// Call endpoint with required userID in context
	req := httptest.NewRequest("POST", "/user/auth/update", strings.NewReader(body))
	ctx := context.WithValue(req.Context(), ctxutil.UserIDKey, int64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.UpdateUserAuthorisation().ServeHTTP(w, req)

	// Code check
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify DB change
	var apiEnabled int64
	row := env.DB.QueryRow("SELECT api_enabled FROM users WHERE email = $1", "example@example.com")
	if err := row.Scan(&apiEnabled); err != nil {
		t.Fatalf("Failed to select value: %v", err)
	}
	if apiEnabled != 0 {
		t.Fatalf("Did not update authorisation for user, expected 0, got %d", apiEnabled)
	}
}

func TestCreateUser_Success(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Domain constructors
	userRepo := user.NewRepository(env.Base, env.Logger)
	userSvc := user.NewService(userRepo, env.Cfg, env.Logger)
	handler := user.NewHandler(userSvc, env.Logger)

	// Create admin user which is required for user creation
	testutil.InsertUser(t, env.DB, "admin@example.com", "x", true, true, true, true)

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
	userRepo := user.NewRepository(env.Base, env.Logger)
	userSvc := user.NewService(userRepo, env.Cfg, env.Logger)
	handler := user.NewHandler(userSvc, env.Logger)

	// Create non-admin user
	testutil.InsertUser(t, env.DB, "example@example.com", "x", true, false, true, true)

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
