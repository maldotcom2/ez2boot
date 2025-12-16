package user_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"ez2boot/internal/shared"
	"ez2boot/internal/testutil"
	"ez2boot/internal/user"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestLogin_Success(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Create user
	email := "example@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, true, false, true, true)

	// Helper checks cookie is returned
	_ = testutil.LoginAndGetCookies(t, env.Router, email, password)
}

func TestLogin_WrongPassword_ReturnsUnauth(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Create user
	email := "example@example.com"

	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, true, false, true, true)

	// Attempt login with wrong password
	loginPayload := user.UserLogin{
		Email:    email,
		Password: "badpassword123",
	}

	loginBody, _ := json.Marshal(loginPayload)
	req := httptest.NewRequest("POST", "/ui/user/login", bytes.NewReader(loginBody))

	// Record the response
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	// Expect 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d, body=%s", w.Code, w.Body.String())
	}

	// Expect no session cookies to be set
	cookies := w.Result().Cookies()
	if len(cookies) != 0 {
		t.Fatalf("expected no cookies, got %d cookies", len(cookies))
	}
}

func TestGetUsers_Success(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Create users
	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	adminHash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, adminHash, true, true, true, true)
	testutil.InsertUser(t, env.DB, "example@example.com", "x", true, false, true, true)
	testutil.InsertUser(t, env.DB, "example2@example.com", "x", true, false, true, true)

	cookies := testutil.LoginAndGetCookies(t, env.Router, adminEmail, adminPassword)

	// Prepare HTTP request to the real route
	req := httptest.NewRequest("GET", "/ui/users", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	// Record the response
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

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

	// Create user
	email := "example@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, true, false, true, true)

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	// Prepare HTTP request to the real route
	req := httptest.NewRequest("GET", "/ui/user/auth", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	// Record the response
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

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

	// Create users
	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	adminHash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, adminHash, true, true, true, true)
	testutil.InsertUser(t, env.DB, "example@example.com", "x", true, false, true, true)

	cookies := testutil.LoginAndGetCookies(t, env.Router, adminEmail, adminPassword)

	// Prepare HTTP request to the real route
	reqPayload := []user.UserAuthRequest{
		{
			UserID:     2,
			IsActive:   true,
			IsAdmin:    false,
			APIEnabled: false,
			UIEnabled:  true,
		},
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("POST", "/ui/user/auth/update", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	// Record the response
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

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

	// Create users
	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	adminHash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, adminHash, true, true, true, true)

	cookies := testutil.LoginAndGetCookies(t, env.Router, adminEmail, adminPassword)

	// Prepare HTTP request to the real route
	reqPayload := user.CreateUserRequest{
		Email:      "test@example.com",
		Password:   "strongpassword123",
		IsActive:   true,
		IsAdmin:    false,
		APIEnabled: true,
		UIEnabled:  true,
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("POST", "/ui/user/new", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	// Record the response
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	// Code check
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify DB row exists
	var email string
	row := env.DB.QueryRow("SELECT email FROM users WHERE email = $1", "test@example.com")
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

	// Create non-admin user
	email := "example@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, true, false, true, true)

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	// Prepare HTTP request to the real route
	reqPayload := user.CreateUserRequest{
		Email:      "test@example.com",
		Password:   "strongpassword123",
		IsActive:   true,
		IsAdmin:    false,
		APIEnabled: true,
		UIEnabled:  true,
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("POST", "/ui/user/new", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	// Record the response
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	// Code check
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d, body=%s", w.Code, w.Body.String())
	}

	// Check if user was created anyway
	var eml string
	row := env.DB.QueryRow("SELECT email FROM users WHERE email = ?", "test@example.com")
	err := row.Scan(&eml)
	if err == nil {
		t.Fatalf("user was created by non-admin: %s", eml)
	}
	if err != sql.ErrNoRows {
		t.Fatalf("unexpected DB error: %v", err)
	}
}
