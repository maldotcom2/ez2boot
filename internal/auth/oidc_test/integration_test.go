package oidc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"ez2boot/internal/auth/oidc"
	"ez2boot/internal/shared"
	"ez2boot/internal/testutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestOidcLogin_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.OidcService.Provider = &testutil.StubOidcProvider{
		AuthCodeURLFunc: func(state string) string {
			return "https://login.microsoftonline.com/authorize?state=" + state
		},
	}

	req := httptest.NewRequest("GET", "/ui/auth/oidc/login", nil)
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("want 302, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify state cookie was set
	cookies := w.Result().Cookies()
	var stateCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "oidc_state" {
			stateCookie = c
			break
		}
	}

	if stateCookie == nil {
		t.Fatal("want oidc_state cookie, got none")
	}

	if stateCookie.Value == "" {
		t.Fatal("want non-empty oidc_state cookie value")
	}

	// Verify redirect location contains the state
	location := w.Header().Get("Location")
	if !strings.Contains(location, stateCookie.Value) {
		t.Fatalf("want redirect to contain state %s, got %s", stateCookie.Value, location)
	}
}

func TestOidcLogin_ProviderNotConfigured(t *testing.T) {
	env := testutil.NewTestEnv(t)
	// Provider is nil by default in test env

	req := httptest.NewRequest("GET", "/ui/auth/oidc/login", nil)
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestOidcCallback_NewUser_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.OidcService.Provider = &testutil.StubOidcProvider{
		ExchangeFunc: func(ctx context.Context, code string) (*oauth2.Token, error) {
			return &oauth2.Token{}, nil
		},
		VerifyIDTokenFunc: func(ctx context.Context, token *oauth2.Token) (map[string]any, error) {
			return map[string]any{
				"preferred_username": "example@example.com",
			}, nil
		},
	}

	req := httptest.NewRequest("GET", "/ui/auth/oidc/callback?code=testcode&state=teststate", nil)
	req.AddCookie(&http.Cookie{Name: "oidc_state", Value: "teststate"})

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("want 302, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify session cookie set
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("want session cookie, got none")
	}
	if sessionCookie.Value == "" {
		t.Fatal("want non-empty session cookie")
	}

	// Verify user was created
	var count int
	err := env.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "example@example.com").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query users: %v", err)
	}
	if count != 1 {
		t.Fatalf("want 1 user, got %d", count)
	}

	// Verify identity provider
	var identityProvider string
	err = env.DB.QueryRow("SELECT identity_provider FROM users WHERE email = $1", "example@example.com").Scan(&identityProvider)
	if err != nil {
		t.Fatalf("failed to query identity provider: %v", err)
	}
	if identityProvider != "oidc" {
		t.Fatalf("want identity_provider=oidc, got %s", identityProvider)
	}
}

func TestOidcCallback_ExistingUser_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	testutil.InsertUser(t, env.DB, "example@example.com", nil, true, false, false, true, "oidc")

	env.OidcService.Provider = &testutil.StubOidcProvider{
		ExchangeFunc: func(ctx context.Context, code string) (*oauth2.Token, error) {
			return &oauth2.Token{}, nil
		},
		VerifyIDTokenFunc: func(ctx context.Context, token *oauth2.Token) (map[string]any, error) {
			return map[string]any{
				"preferred_username": "example@example.com",
			}, nil
		},
	}

	req := httptest.NewRequest("GET", "/ui/auth/oidc/callback?code=testcode&state=teststate", nil)
	req.AddCookie(&http.Cookie{Name: "oidc_state", Value: "teststate"})

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("want 302, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify session cookie set
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("want session cookie, got none")
	}

	// Verify no duplicate user created
	var count int
	err := env.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "example@example.com").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query users: %v", err)
	}
	if count != 1 {
		t.Fatalf("want 1 user, got %d", count)
	}
}

func TestOidcCallback_WrongIdentityProvider(t *testing.T) {
	env := testutil.NewTestEnv(t)

	// Insert a local user with the same email
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, "example@example.com", &hash, true, false, false, true, "local")

	env.OidcService.Provider = &testutil.StubOidcProvider{
		ExchangeFunc: func(ctx context.Context, code string) (*oauth2.Token, error) {
			return &oauth2.Token{}, nil
		},
		VerifyIDTokenFunc: func(ctx context.Context, token *oauth2.Token) (map[string]any, error) {
			return map[string]any{
				"preferred_username": "example@example.com",
			}, nil
		},
	}

	req := httptest.NewRequest("GET", "/ui/auth/oidc/callback?code=testcode&state=teststate", nil)
	req.AddCookie(&http.Cookie{Name: "oidc_state", Value: "teststate"})

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify the local user was not given a session
	var count int
	err := env.DB.QueryRow("SELECT COUNT(*) FROM user_sessions WHERE user_id = (SELECT id FROM users WHERE email = $1)", "example@example.com").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query sessions: %v", err)
	}
	if count != 0 {
		t.Fatalf("want 0 sessions, got %d", count)
	}
}

func TestOidcCallback_StateMismatch(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.OidcService.Provider = &testutil.StubOidcProvider{}

	req := httptest.NewRequest("GET", "/ui/auth/oidc/callback?code=testcode&state=wrongstate", nil)
	req.AddCookie(&http.Cookie{Name: "oidc_state", Value: "correctstate"})

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestOidcCallback_ExchangeFailure(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.OidcService.Provider = &testutil.StubOidcProvider{
		ExchangeFunc: func(ctx context.Context, code string) (*oauth2.Token, error) {
			return nil, errors.New("exchange failed")
		},
	}

	req := httptest.NewRequest("GET", "/ui/auth/oidc/callback?code=testcode&state=teststate", nil)
	req.AddCookie(&http.Cookie{Name: "oidc_state", Value: "teststate"})

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestOidcCallback_VerifyTokenFailure(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.OidcService.Provider = &testutil.StubOidcProvider{
		ExchangeFunc: func(ctx context.Context, code string) (*oauth2.Token, error) {
			return &oauth2.Token{}, nil
		},
		VerifyIDTokenFunc: func(ctx context.Context, token *oauth2.Token) (map[string]any, error) {
			return nil, errors.New("token verification failed")
		},
	}

	req := httptest.NewRequest("GET", "/ui/auth/oidc/callback?code=testcode&state=teststate", nil)
	req.AddCookie(&http.Cookie{Name: "oidc_state", Value: "teststate"})

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestOidcCallback_NoEmailInClaims(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.OidcService.Provider = &testutil.StubOidcProvider{
		ExchangeFunc: func(ctx context.Context, code string) (*oauth2.Token, error) {
			return &oauth2.Token{}, nil
		},
		VerifyIDTokenFunc: func(ctx context.Context, token *oauth2.Token) (map[string]any, error) {
			return map[string]any{
				"sub": "12345",
			}, nil
		},
	}

	req := httptest.NewRequest("GET", "/ui/auth/oidc/callback?code=testcode&state=teststate", nil)
	req.AddCookie(&http.Cookie{Name: "oidc_state", Value: "teststate"})

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestOidcCallback_InactiveUser(t *testing.T) {
	env := testutil.NewTestEnv(t)

	testutil.InsertUser(t, env.DB, "example@example.com", nil, false, false, false, true, "oidc")

	env.OidcService.Provider = &testutil.StubOidcProvider{
		ExchangeFunc: func(ctx context.Context, code string) (*oauth2.Token, error) {
			return &oauth2.Token{}, nil
		},
		VerifyIDTokenFunc: func(ctx context.Context, token *oauth2.Token) (map[string]any, error) {
			return map[string]any{
				"preferred_username": "example@example.com",
			}, nil
		},
	}

	req := httptest.NewRequest("GET", "/ui/auth/oidc/callback?code=testcode&state=teststate", nil)
	req.AddCookie(&http.Cookie{Name: "oidc_state", Value: "teststate"})

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestGetOidcConfig_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, &hash, true, true, true, true, "local")
	testutil.InsertOidcConfig(t, env.DB, env.Encryptor, "https://login.microsoftonline.com/test/v2.0", "test-client-id", "test-secret", "http://localhost:8000/ui/auth/oidc/callback")

	cookies := testutil.LoginAndGetCookies(t, env.Router, adminEmail, adminPassword)

	req := httptest.NewRequest("GET", "/ui/auth/oidc", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp shared.ApiResponse[any]
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Fatal("want success=true, got false")
	}

	if resp.Data == nil {
		t.Fatal("want data, got nil")
	}
}

func TestGetOidcConfig_NotFound(t *testing.T) {
	env := testutil.NewTestEnv(t)

	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, &hash, true, true, true, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, adminEmail, adminPassword)

	req := httptest.NewRequest("GET", "/ui/auth/oidc", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp shared.ApiResponse[any]
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Fatal("want success=true, got false")
	}

	if resp.Data != nil {
		t.Fatal("want data=nil, got non-nil")
	}
}

func TestSetOidcConfig_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, &hash, true, true, true, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, adminEmail, adminPassword)

	payload := oidc.OidcConfigRequest{
		IssuerURL:    "https://login.microsoftonline.com/test/v2.0",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		AppURL:       "http://localhost:8000",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/ui/auth/oidc", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify config was saved
	var issuerURL string
	err := env.DB.QueryRow("SELECT issuer_url FROM oidc_config WHERE id = 1").Scan(&issuerURL)
	if err != nil {
		t.Fatalf("failed to query oidc config: %v", err)
	}
	if issuerURL != "https://login.microsoftonline.com/test/v2.0" {
		t.Fatalf("want issuer_url=https://login.microsoftonline.com/test/v2.0, got %s", issuerURL)
	}
}

func TestSetOidcConfig_NotAdmin_ReturnsForbidden(t *testing.T) {
	env := testutil.NewTestEnv(t)

	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, &hash, true, true, true, true, "local")
	testutil.InsertUser(t, env.DB, "example@example.com", &hash, true, false, false, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, "example@example.com", adminPassword)

	payload := oidc.OidcConfigRequest{
		IssuerURL:    "https://login.microsoftonline.com/test/v2.0",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		AppURL:       "http://localhost:8000",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/ui/auth/oidc", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestDeleteOidcConfig_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, &hash, true, true, true, true, "local")
	testutil.InsertOidcConfig(t, env.DB, env.Encryptor, "https://login.microsoftonline.com/test/v2.0", "test-client-id", "test-secret", "http://localhost:8000")

	cookies := testutil.LoginAndGetCookies(t, env.Router, adminEmail, adminPassword)

	req := httptest.NewRequest("DELETE", "/ui/auth/oidc", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify config was deleted
	var count int
	err := env.DB.QueryRow("SELECT COUNT(*) FROM oidc_config").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query oidc config: %v", err)
	}
	if count != 0 {
		t.Fatalf("want 0 configs, got %d", count)
	}

	// Verify provider was cleared
	if env.OidcService.Provider != nil {
		t.Fatal("want provider=nil after delete")
	}
}

func TestDeleteOidcConfig_NotAdmin_returnsForbidden(t *testing.T) {
	env := testutil.NewTestEnv(t)

	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, &hash, true, true, true, true, "local")
	testutil.InsertUser(t, env.DB, "example@example.com", &hash, true, false, false, true, "local")
	testutil.InsertOidcConfig(t, env.DB, env.Encryptor, "https://login.microsoftonline.com/test/v2.0", "test-client-id", "test-secret", "http://localhost:8000")

	cookies := testutil.LoginAndGetCookies(t, env.Router, "example@example.com", adminPassword)

	req := httptest.NewRequest("DELETE", "/ui/auth/oidc", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify config was not deleted
	var count int
	err := env.DB.QueryRow("SELECT COUNT(*) FROM oidc_config").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query oidc config: %v", err)
	}
	if count != 1 {
		t.Fatalf("want 1 config, got %d", count)
	}
}

func TestHasOidc_True(t *testing.T) {
	env := testutil.NewTestEnv(t)

	env.OidcService.Provider = &testutil.StubOidcProvider{}

	req := httptest.NewRequest("GET", "/ui/auth/oidc/status", nil)
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp shared.ApiResponse[oidc.HasOidcRespose]
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Fatal("want success=true, got false")
	}

	if !resp.Data.HasOidc {
		t.Fatal("want has_oidc=true, got false")
	}
}

func TestHasOidc_False(t *testing.T) {
	env := testutil.NewTestEnv(t)

	req := httptest.NewRequest("GET", "/ui/auth/oidc/status", nil)
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp shared.ApiResponse[oidc.HasOidcRespose]
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Fatal("want success=true, got false")
	}

	if resp.Data.HasOidc {
		t.Fatal("want has_oidc=false, got true")
	}
}
