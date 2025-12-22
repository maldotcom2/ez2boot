package user_test

import (
	"bytes"
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

	// Helper confirms login
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
		t.Fatalf("want 401, got %d, body=%s", w.Code, w.Body.String())
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
	testutil.InsertUser(t, env.DB, email, hash, true, false, true, true)

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
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
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
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
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
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify DB change
	var apiEnabled int64
	err := env.DB.QueryRow("SELECT api_enabled FROM users WHERE email = $1", "example@example.com").Scan(&apiEnabled)
	if err != nil {
		t.Fatalf("Failed to select value: %v", err)
	}
	if apiEnabled != 0 {
		t.Fatalf("Did not update authorisation for user, want 0, got %d", apiEnabled)
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
		t.Fatalf("want 201, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify DB row exists
	var count int64
	err := env.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "test@example.com").Scan(&count)
	if err != nil {
		t.Fatalf("user not inserted: %v", err)
	}
}

func TestCreateUser_NotAdmin_ReturnsForbidden(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Create non-admin user
	email := "example@example.com"
	password := "testpassword123"
	newUser := "test@example.com"
	newUserPassword := "strongpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, true, false, true, true)

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	// Prepare HTTP request to the real route
	reqPayload := user.CreateUserRequest{
		Email:      newUser,
		Password:   newUserPassword,
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
		t.Fatalf("want 403, got %d, body=%s", w.Code, w.Body.String())
	}

	// Check if user was created anyway
	var count int64
	err := env.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", newUser).Scan(&count)
	if err != nil {
		t.Fatalf("db query failed: %v", err)
	}

	if count != 0 {
		t.Fatalf("row count for non-admin create user want: 0, found %d rows", count)
	}
}

func TestDeleteUser_Success(t *testing.T) {
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
	reqPayload := user.DeleteUserRequest{
		UserID: 2,
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("DELETE", "/ui/user/delete", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	// Record the response
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	// Code check
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify DB row removed
	var count int64
	err := env.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "example@example.com").Scan(&count)
	if err != nil {
		t.Fatalf("db query failed: %v", err)
	}

	if count != 0 {
		t.Fatalf("row count for deleted user want: 0, found %d rows", count)
	}
}

func TestCreateFirstTimeUser_Success(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Prepare HTTP request to the real route
	reqPayload := user.CreateUserRequest{
		Email:    "admin@example.com",
		Password: "testpassword123",
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("POST", "/ui/setup", bytes.NewReader(body))

	// Record the response
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	// Code check
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify DB entry
	var u user.User
	row := env.DB.QueryRow("SELECT email, is_active, is_admin, ui_enabled, api_enabled FROM users WHERE email = $1", "admin@example.com")
	if err := row.Scan(&u.Email, &u.IsActive, &u.IsAdmin, &u.UIEnabled, &u.APIEnabled); err != nil {
		t.Fatalf("user not inserted: %v", err)
	}
	if u.IsActive != true {
		t.Fatalf("IsActive want: true, got: %v", u.IsActive)
	}
	if u.IsAdmin != true {
		t.Fatalf("IsAdmin want: true, got: %v", u.IsAdmin)
	}
	if u.UIEnabled != true {
		t.Fatalf("UIEnabled want: true, got: %v", u.UIEnabled)
	}
	if u.APIEnabled != true {
		t.Fatalf("APIEnabled want: true, got: %v", u.APIEnabled)
	}
}

func TestCreateFirstTimeUser_SecondUser_ReturnsForbidden(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Prepare HTTP request to the real route
	firstReqPayload := user.CreateUserRequest{
		Email:    "admin@example.com",
		Password: "testpassword123",
	}

	body, _ := json.Marshal(firstReqPayload)
	req := httptest.NewRequest("POST", "/ui/setup", bytes.NewReader(body))

	// Record the response
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	// Code check
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", w.Code, w.Body.String())
	}

	// Prepare HTTP request to the real route
	secondReqPayload := user.CreateUserRequest{
		Email:    "admin2@example.com",
		Password: "testpassword456",
	}

	body, _ = json.Marshal(secondReqPayload)
	req = httptest.NewRequest("POST", "/ui/setup", bytes.NewReader(body))

	// Record the response
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	// Code check
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify no second user
	var count int
	err := env.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "admin2@example.com").Scan(&count)
	if err != nil {
		t.Fatalf("db query failed: %v", err)
	}

	if count != 0 {
		t.Fatalf("row count for second user want: 0, found %d rows", count)
	}
}

func TestChangePassword_Success(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Create users
	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	adminNewPassword := "testpassword456"
	adminHash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, adminHash, true, true, true, true)

	cookies := testutil.LoginAndGetCookies(t, env.Router, adminEmail, adminPassword)

	// Prepare HTTP request to the real route
	reqPayload := user.ChangePasswordRequest{
		OldPassword: adminPassword,
		NewPassword: adminNewPassword,
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("PUT", "/ui/user/changepassword", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	// Record the response
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	// Code check
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// Attempt login with new password
	_ = testutil.LoginAndGetCookies(t, env.Router, adminEmail, adminNewPassword)
}
