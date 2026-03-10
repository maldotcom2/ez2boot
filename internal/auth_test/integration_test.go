package auth_test

import (
	"bytes"
	"encoding/json"
	"ez2boot/internal/auth"
	"ez2boot/internal/testutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogin_Success(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Create user
	email := "example@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, true, false, true, true, "local")

	// Helper confirms login
	_ = testutil.LoginAndGetCookies(t, env.Router, email, password)
}

func TestLogin_WrongPassword_ReturnsUnauth(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Create user
	email := "example@example.com"

	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, true, false, true, true, "local")

	// Attempt login with wrong password
	loginPayload := auth.UserLogin{
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
		t.Fatalf("want 401, got %d, body=%s", w.Code, w.Body.String())
	}

	// Expect no session cookies to be set
	cookies := w.Result().Cookies()
	if len(cookies) != 0 {
		t.Fatalf("want no cookies, got %d cookies", len(cookies))
	}
}

func TestLogin_Inactive_ReturnsForbidden(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Create user
	email := "example@example.com"

	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, false, false, true, true, "local")

	// Attempt login
	loginPayload := auth.UserLogin{
		Email:    email,
		Password: "testpassword123",
	}

	loginBody, _ := json.Marshal(loginPayload)
	req := httptest.NewRequest("POST", "/ui/user/login", bytes.NewReader(loginBody))

	// Record the response
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	// Expect 403 Forbidden
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d, body=%s", w.Code, w.Body.String())
	}

	// Expect no session cookies to be set
	cookies := w.Result().Cookies()
	if len(cookies) != 0 {
		t.Fatalf("want no cookies, got %d cookies", len(cookies))
	}
}

func TestLogin_UIblocked_ReturnsForbidden(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Create user
	email := "example@example.com"

	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, true, false, true, false, "local")

	// Attempt login
	loginPayload := auth.UserLogin{
		Email:    email,
		Password: "testpassword123",
	}

	loginBody, _ := json.Marshal(loginPayload)
	req := httptest.NewRequest("POST", "/ui/user/login", bytes.NewReader(loginBody))

	// Record the response
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	// Expect 403 Forbidden
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d, body=%s", w.Code, w.Body.String())
	}

	// Expect no session cookies to be set
	cookies := w.Result().Cookies()
	if len(cookies) != 0 {
		t.Fatalf("want no cookies, got %d cookies", len(cookies))
	}
}

func TestLogout_Success(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Create user
	email := "example@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, true, false, true, true, "local")

	// Helper confirms login
	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	// Prepare HTTP request to the real route
	req := httptest.NewRequest("POST", "/ui/user/logout", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	// Record the response
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	// Check HTTP status code
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify DB change
	var count int64
	err := env.DB.QueryRow("SELECT COUNT(*) FROM user_sessions WHERE user_id = $1", 1).Scan(&count)
	if err != nil {
		t.Fatalf("db query failed: %v", err)
	}

	if count != 0 {
		t.Fatalf("row count for user session want: 0, found %d rows", count)
	}
}
