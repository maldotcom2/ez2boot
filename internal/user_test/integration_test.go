package user_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"ez2boot/internal/auth"
	"ez2boot/internal/shared"
	"ez2boot/internal/testutil"
	"ez2boot/internal/user"
	"ez2boot/internal/util"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pquerna/otp/totp"
)

func TestGetUsers_Success(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Create users
	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	adminHash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, &adminHash, true, true, true, true, "local")
	testutil.InsertUser(t, env.DB, "example@example.com", nil, true, false, true, true, "local")
	testutil.InsertUser(t, env.DB, "example2@example.com", nil, true, false, true, true, "local")

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
	var got shared.ApiResponse[[]user.GetUsersResponse]
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Zero last login value for test comparison
	for i := range got.Data {
		got.Data[i].LastLogin = nil
	}

	// Expected API response
	want := shared.ApiResponse[[]user.GetUsersResponse]{
		Success: true,
		Data: []user.GetUsersResponse{
			{
				UserID:           1,
				Email:            "admin@example.com",
				IsActive:         true,
				IsAdmin:          true,
				APIEnabled:       true,
				UIEnabled:        true,
				IdentityProvider: "local",
				LastLogin:        nil,
			},
			{
				UserID:           2,
				Email:            "example@example.com",
				IsActive:         true,
				IsAdmin:          false,
				APIEnabled:       true,
				UIEnabled:        true,
				IdentityProvider: "local",
				LastLogin:        nil,
			},
			{
				UserID:           3,
				Email:            "example2@example.com",
				IsActive:         true,
				IsAdmin:          false,
				APIEnabled:       true,
				UIEnabled:        true,
				IdentityProvider: "local",
				LastLogin:        nil,
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
	testutil.InsertUser(t, env.DB, email, &hash, true, false, true, true, "local")

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

	var got shared.ApiResponse[user.UserAuthResponse]

	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Expected API response
	want := shared.ApiResponse[user.UserAuthResponse]{
		Success: true,
		Data: user.UserAuthResponse{
			UserID:           1,
			Email:            "example@example.com",
			IsActive:         true,
			IsAdmin:          false,
			APIEnabled:       true,
			UIEnabled:        true,
			IdentityProvider: "local",
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
	testutil.InsertUser(t, env.DB, adminEmail, &adminHash, true, true, true, true, "local")
	testutil.InsertUser(t, env.DB, "example@example.com", nil, true, false, true, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, adminEmail, adminPassword)

	// Prepare HTTP request to the real route
	reqPayload := []user.UserAuthResponse{
		{
			UserID:     2,
			IsActive:   true,
			IsAdmin:    false,
			APIEnabled: false,
			UIEnabled:  true,
		},
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("PUT", "/ui/user/auth", bytes.NewReader(body))
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

func TestUpdateUserAuthorisation_ExternalUserAPIAccess_ReturnsBad(t *testing.T) {
	env := testutil.NewTestEnv(t)

	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	adminHash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, &adminHash, true, true, true, true, "local")
	testutil.InsertUser(t, env.DB, "ldapuser@example.com", nil, true, false, false, true, "ldap")

	cookies := testutil.LoginAndGetCookies(t, env.Router, adminEmail, adminPassword)

	reqPayload := []user.UpdateUserRequest{
		{
			UserID:     2,
			IsActive:   true,
			IsAdmin:    false,
			APIEnabled: true,
			UIEnabled:  true,
		},
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("PUT", "/ui/user/auth", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify DB unchanged
	var apiEnabled int64
	err := env.DB.QueryRow("SELECT api_enabled FROM users WHERE email = $1", "ldapuser@example.com").Scan(&apiEnabled)
	if err != nil {
		t.Fatalf("Failed to select value: %v", err)
	}
	if apiEnabled != 0 {
		t.Fatalf("API access was granted to external user, want 0, got %d", apiEnabled)
	}
}

func TestCreateUser_Success(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Create users
	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	adminHash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, &adminHash, true, true, true, true, "local")

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
	req := httptest.NewRequest("POST", "/ui/user", bytes.NewReader(body))
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
	testutil.InsertUser(t, env.DB, email, &hash, true, false, true, true, "local")

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
	req := httptest.NewRequest("POST", "/ui/user", bytes.NewReader(body))
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
	testutil.InsertUser(t, env.DB, adminEmail, &adminHash, true, true, true, true, "local")
	testutil.InsertUser(t, env.DB, "example@example.com", nil, true, false, true, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, adminEmail, adminPassword)

	// Prepare HTTP request to the real route
	reqPayload := user.DeleteUserRequest{
		UserID: 2,
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("DELETE", "/ui/user", bytes.NewReader(body))
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
	var u user.GetUsersResponse
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
	testutil.InsertUser(t, env.DB, adminEmail, &adminHash, true, true, true, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, adminEmail, adminPassword)

	// Prepare HTTP request to the real route
	reqPayload := user.ChangePasswordRequest{
		CurrentPassword: adminPassword,
		NewPassword:     adminNewPassword,
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("PUT", "/ui/user/password", bytes.NewReader(body))
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

func TestEnrolMFA_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "user@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, &hash, true, true, false, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	req := httptest.NewRequest("POST", "/ui/user/mfa", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp shared.ApiResponse[string]
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Fatalf("want success=true, got false, error=%s", resp.Error)
	}

	// Verify base64 QR code is returned
	if resp.Data == "" {
		t.Fatal("want base64 QR code, got empty string")
	}

	// Verify it's valid base64
	_, err := base64.StdEncoding.DecodeString(resp.Data)
	if err != nil {
		t.Fatalf("want valid base64, got error: %v", err)
	}

	// Verify secret was stored in DB
	var secret *string
	env.DB.QueryRow("SELECT mfa_secret FROM users WHERE email = $1", email).Scan(&secret)
	if secret == nil {
		t.Fatal("want mfa_secret to be set, got nil")
	}

	// Verify mfa_confirmed is still 0
	var mfaConfirmed int64
	env.DB.QueryRow("SELECT mfa_confirmed FROM users WHERE email = $1", email).Scan(&mfaConfirmed)
	if mfaConfirmed != 0 {
		t.Fatalf("want mfa_confirmed=0, got %d", mfaConfirmed)
	}
}

func TestEnrolMFA_OIDCUser(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "oidcuser@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, &hash, true, true, true, true, "oidc")

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	req := httptest.NewRequest("POST", "/ui/user/mfa", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp shared.ApiResponse[any]
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Success {
		t.Fatal("want success=false, got true")
	}
}

func TestEnrolMFA_Unauthenticated(t *testing.T) {
	env := testutil.NewTestEnv(t)

	req := httptest.NewRequest("POST", "/ui/user/mfa", nil)

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestEnrolMFA_ReenrolOverwritesSecret(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "user@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, &hash, true, true, false, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	// First enrolment
	req := httptest.NewRequest("POST", "/ui/user/mfa", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	var firstSecret *string // Secrets are nullable
	env.DB.QueryRow("SELECT mfa_secret FROM users WHERE email = $1", email).Scan(&firstSecret)

	// Second enrolment
	req = httptest.NewRequest("POST", "/ui/user/mfa", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	var secondSecret *string
	env.DB.QueryRow("SELECT mfa_secret FROM users WHERE email = $1", email).Scan(&secondSecret)

	// Secrets should be different
	if *firstSecret == *secondSecret {
		t.Fatal("want new secret on re-enrol, got same secret")
	}
}

func TestConfirmMFA_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "user@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, &hash, true, true, false, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	// Enrol first to get a secret stored
	req := httptest.NewRequest("POST", "/ui/user/mfa", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("enrol failed: want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// Get the stored secret from DB to generate a valid code
	var secret string
	env.DB.QueryRow("SELECT mfa_secret FROM users WHERE email = $1", email).Scan(&secret)

	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("failed to generate TOTP code: %v", err)
	}

	reqPayload := user.MFARequest{
		Code: code,
	}

	body, _ := json.Marshal(reqPayload)
	req = httptest.NewRequest("POST", "/ui/user/mfa/confirm", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp shared.ApiResponse[any]
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Fatalf("want success=true, got false, error=%s", resp.Error)
	}

	// Verify mfa_confirmed is now 1 in DB
	var mfaConfirmed int64
	env.DB.QueryRow("SELECT mfa_confirmed FROM users WHERE email = $1", email).Scan(&mfaConfirmed)
	if mfaConfirmed != 1 {
		t.Fatalf("want mfa_confirmed=1, got %d", mfaConfirmed)
	}
}

func TestConfirmMFA_IncorrectCode(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "user@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, &hash, true, true, false, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	// Enrol first
	req := httptest.NewRequest("POST", "/ui/user/mfa", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	reqPayload := user.MFARequest{
		Code: "000000",
	}

	body, _ := json.Marshal(reqPayload)
	req = httptest.NewRequest("POST", "/ui/user/mfa/confirm", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestConfirmMFA_NotEnrolled(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "user@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, &hash, true, true, false, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	// Skip enrolment, go straight to confirm
	reqPayload := user.MFARequest{
		Code: "123456",
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("POST", "/ui/user/mfa/confirm", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestConfirmMFA_AlreadyConfirmed(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "user@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, &hash, true, true, false, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	// Enrol
	req := httptest.NewRequest("POST", "/ui/user/mfa", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	var secret string
	env.DB.QueryRow("SELECT mfa_secret FROM users WHERE email = $1", email).Scan(&secret)

	code, _ := totp.GenerateCode(secret, time.Now())

	// First confirm
	reqPayload := user.MFARequest{
		Code: code,
	}

	body, _ := json.Marshal(reqPayload)
	req = httptest.NewRequest("POST", "/ui/user/mfa/confirm", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	// Second confirm with a new code (can't reuse same code due to cache)
	code2, _ := totp.GenerateCode(secret, time.Now().Add(31*time.Second))

	reqPayload = user.MFARequest{
		Code: code2,
	}

	body, _ = json.Marshal(reqPayload)
	req = httptest.NewRequest("POST", "/ui/user/mfa/confirm", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestConfirmMFA_ReplayAttack(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "user@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, &hash, true, true, false, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	// Enrol
	req := httptest.NewRequest("POST", "/ui/user/mfa", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	var secret string
	env.DB.QueryRow("SELECT mfa_secret FROM users WHERE email = $1", email).Scan(&secret)

	code, _ := totp.GenerateCode(secret, time.Now())

	// First use of code
	reqPayload := user.MFARequest{
		Code: code,
	}

	body, _ := json.Marshal(reqPayload)
	req = httptest.NewRequest("POST", "/ui/user/mfa/confirm", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("first confirm: want 200, got %d", w.Code)
	}

	// Replay same code — reset mfa_confirmed to 0 to isolate the cache check
	env.DB.Exec("UPDATE users SET mfa_confirmed = 0 WHERE email = $1", email)

	req = httptest.NewRequest("POST", "/ui/user/mfa/confirm", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("replay: want 400, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestConfirmMFA_Unauthenticated(t *testing.T) {
	env := testutil.NewTestEnv(t)

	reqPayload := user.MFARequest{
		Code: "123456",
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("POST", "/ui/user/mfa/confirm", bytes.NewReader(body))

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestDeleteMFA_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "user@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, &hash, true, true, false, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	// Enrol and confirm MFA
	req := httptest.NewRequest("POST", "/ui/user/mfa", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	var secret string
	env.DB.QueryRow("SELECT mfa_secret FROM users WHERE email = $1", email).Scan(&secret)

	confirmCode, _ := totp.GenerateCode(secret, time.Now())

	reqPayload := user.MFARequest{
		Code: confirmCode,
	}

	body, _ := json.Marshal(reqPayload)
	req = httptest.NewRequest("POST", "/ui/user/mfa/confirm", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("confirm failed: want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// Delete MFA with a new code
	deleteCode, _ := totp.GenerateCode(secret, time.Now().Add(31*time.Second))

	reqPayload = user.MFARequest{
		Code: deleteCode,
	}

	body, _ = json.Marshal(reqPayload)
	req = httptest.NewRequest("POST", "/ui/user/mfa/delete", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify secret is null in DB
	var dbSecret *string
	env.DB.QueryRow("SELECT mfa_secret FROM users WHERE email = $1", email).Scan(&dbSecret)
	if dbSecret != nil {
		t.Fatal("want mfa_secret to be nil, got value")
	}

	// Verify mfa_confirmed is 0
	var mfaConfirmed int64
	env.DB.QueryRow("SELECT mfa_Confirmed FROM users WHERE email = $1", email).Scan(&mfaConfirmed)
	if mfaConfirmed != 0 {
		t.Fatalf("want mfa_confirmed=0, got %d", mfaConfirmed)
	}
}

func TestDeleteMFA_Unauthenticated(t *testing.T) {
	env := testutil.NewTestEnv(t)

	reqPayload := user.MFARequest{
		Code: "123456",
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("POST", "/ui/user/mfa/delete", bytes.NewReader(body))

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestVerifyMFA_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "user@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, &hash, true, true, false, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	// Enrol
	req := httptest.NewRequest("POST", "/ui/user/mfa", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	var secret string
	env.DB.QueryRow("SELECT mfa_secret FROM users WHERE email = $1", email).Scan(&secret)

	// Confirm
	confirmCode, _ := totp.GenerateCode(secret, time.Now())

	reqPayload := user.MFARequest{
		Code: confirmCode,
	}

	body, _ := json.Marshal(reqPayload)
	req = httptest.NewRequest("POST", "/ui/user/mfa/confirm", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("confirm failed: want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// Login again to trigger MFA flow
	loginPayload := auth.UserLogin{
		Email:    email,
		Password: password,
	}

	body, _ = json.Marshal(loginPayload)
	req = httptest.NewRequest("POST", "/ui/auth/login", bytes.NewReader(body))
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var mfaPendingCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "mfa_pending" {
			mfaPendingCookie = c
		}
	}

	if mfaPendingCookie == nil {
		t.Fatal("want mfa_pending cookie, got none")
	}

	// Verify
	verifyCode, _ := totp.GenerateCode(secret, time.Now().Add(31*time.Second))

	reqPayload = user.MFARequest{
		Code: verifyCode,
	}

	body, _ = json.Marshal(reqPayload)
	req = httptest.NewRequest("POST", "/ui/user/mfa/verify", bytes.NewReader(body))
	req.AddCookie(mfaPendingCookie)
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify session cookie issued
	var sessionCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "session" {
			sessionCookie = c
		}
	}

	if sessionCookie == nil {
		t.Fatal("want session cookie, got none")
	}

	// Verify pending session deleted from DB
	var count int
	env.DB.QueryRow("SELECT COUNT(*) FROM mfa_pending_sessions WHERE user_id = (SELECT id FROM users WHERE email = $1)", email).Scan(&count)
	if count != 0 {
		t.Fatalf("want mfa_pending_session deleted, got %d rows", count)
	}
}

func TestVerifyMFA_NoPendingCookie(t *testing.T) {
	env := testutil.NewTestEnv(t)

	reqPayload := user.MFARequest{
		Code: "123456",
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("POST", "/ui/user/mfa/verify", bytes.NewReader(body))

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestVerifyMFA_ExpiredSession(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "user@example.com"
	//password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, &hash, true, true, false, true, "local")

	// Insert expired pending session directly into DB
	expiredToken, _ := util.GenerateRandomString(32)
	expiredHash := util.HashToken(expiredToken)
	expiry := time.Now().Add(-1 * time.Minute).Unix()

	var userID int64
	env.DB.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	env.DB.Exec("INSERT INTO mfa_pending_sessions (token_hash, user_id, session_expiry) VALUES ($1, $2, $3)", expiredHash, userID, expiry)

	reqPayload := user.MFARequest{
		Code: "123456", // Expiry is checked before code, code doesn't matter here
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("POST", "/ui/user/mfa/verify", bytes.NewReader(body))
	req.AddCookie(&http.Cookie{
		Name:  "mfa_pending",
		Value: expiredToken,
	})

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestVerifyMFA_IncorrectCode(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "user@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, &hash, true, true, false, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	// Enrol
	req := httptest.NewRequest("POST", "/ui/user/mfa", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	var secret string
	env.DB.QueryRow("SELECT mfa_secret FROM users WHERE email = $1", email).Scan(&secret)

	// Confirm
	confirmCode, _ := totp.GenerateCode(secret, time.Now())

	reqPayload := user.MFARequest{
		Code: confirmCode,
	}

	body, _ := json.Marshal(reqPayload)
	req = httptest.NewRequest("POST", "/ui/user/mfa/confirm", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	// Login again to get mfa_pending cookie
	loginPayload := auth.UserLogin{
		Email:    email,
		Password: password,
	}

	body, _ = json.Marshal(loginPayload)
	req = httptest.NewRequest("POST", "/ui/auth/login", bytes.NewReader(body))
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	var mfaPendingCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "mfa_pending" {
			mfaPendingCookie = c
		}
	}

	reqPayload = user.MFARequest{
		Code: "000000",
	}

	body, _ = json.Marshal(reqPayload)
	req = httptest.NewRequest("POST", "/ui/user/mfa/verify", bytes.NewReader(body))
	req.AddCookie(mfaPendingCookie)
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d, body=%s", w.Code, w.Body.String())
	}
}
