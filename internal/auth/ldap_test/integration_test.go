package ldap_test

import (
	"bytes"
	"encoding/json"
	"ez2boot/internal/auth/ldap"
	"ez2boot/internal/shared"
	"ez2boot/internal/testutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetLdapConfig_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	// Create admin user
	email := "admin@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, true, true, false, true, "local")

	// Insert LDAP config
	testutil.InsertLdapConfig(t, env.DB, env.Encryptor, "AD01", 389, "dc=ez2boot,dc=org", "CN=ldap.svc,CN=Users,DC=ez2boot,DC=org", "password", false, false)

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	req := httptest.NewRequest("GET", "/ui/auth/ldap", nil)
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

	var got shared.ApiResponse[ldap.LdapConfigResponse]
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	want := shared.ApiResponse[ldap.LdapConfigResponse]{
		Success: true,
		Data: ldap.LdapConfigResponse{
			Host:          "AD01",
			Port:          389,
			BaseDN:        "dc=ez2boot,dc=org",
			BindDN:        "CN=ldap.svc,CN=Users,DC=ez2boot,DC=org",
			BindPassword:  "", // Password is nulled in response
			UseSSL:        false,
			SkipTLSVerify: false,
		},
		Error: "",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("body mismatch\n got:  %#v\n want: %#v", got, want)
	}

	// Verify password was stored and is decryptable
	var encPassword []byte
	err := env.DB.QueryRow("SELECT bind_password FROM ldap_config WHERE id = 1").Scan(&encPassword)
	if err != nil {
		t.Fatalf("failed to query bind password: %v", err)
	}

	decrypted, err := env.Encryptor.Decrypt(encPassword)
	if err != nil {
		t.Fatalf("failed to decrypt bind password: %v", err)
	}

	if string(decrypted) != "password" {
		t.Fatalf("bind password mismatch, want 'password', got '%s'", string(decrypted))
	}
}

// Success and nil data required for UI functionality
func TestGetLdapConfig_OK(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "admin@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, true, true, false, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	req := httptest.NewRequest("GET", "/ui/auth/ldap", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var got shared.ApiResponse[any]
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !got.Success {
		t.Fatalf("want success=true, got false")
	}

	if got.Data != nil {
		t.Fatalf("want data=nil, got %v", got.Data)
	}
}

func TestSetLdapConfig_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "admin@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, true, true, false, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	reqPayload := ldap.LdapConfigRequest{
		Host:          "AD01",
		Port:          389,
		BaseDN:        "dc=ez2boot,dc=org",
		BindDN:        "CN=ldap.svc,CN=Users,DC=ez2boot,DC=org",
		BindPassword:  "password",
		UseSSL:        false,
		SkipTLSVerify: false,
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("POST", "/ui/auth/ldap", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var got shared.ApiResponse[any]
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !got.Success {
		t.Fatalf("want success=true, got false")
	}

	// Verify DB state
	var host string
	var port int64
	var baseDN, bindDN string
	var useSSL, skipTLSVerify bool
	var encPassword []byte

	err := env.DB.QueryRow("SELECT host, port, base_dn, bind_dn, bind_password, use_ssl, skip_tls_verify FROM ldap_config WHERE id = 1").
		Scan(&host, &port, &baseDN, &bindDN, &encPassword, &useSSL, &skipTLSVerify)
	if err != nil {
		t.Fatalf("failed to query ldap config: %v", err)
	}

	if host != "AD01" {
		t.Fatalf("host mismatch, want 'AD01', got '%s'", host)
	}
	if port != 389 {
		t.Fatalf("port mismatch, want 389, got %d", port)
	}
	if baseDN != "dc=ez2boot,dc=org" {
		t.Fatalf("base_dn mismatch, want 'dc=ez2boot,dc=org', got '%s'", baseDN)
	}
	if bindDN != "CN=ldap.svc,CN=Users,DC=ez2boot,DC=org" {
		t.Fatalf("bind_dn mismatch, want 'CN=ldap.svc,CN=Users,DC=ez2boot,DC=org', got '%s'", bindDN)
	}

	// Verify password was encrypted and is decryptable
	decrypted, err := env.Encryptor.Decrypt(encPassword)
	if err != nil {
		t.Fatalf("failed to decrypt bind password: %v", err)
	}
	if string(decrypted) != "password" {
		t.Fatalf("bind password mismatch, want 'password', got '%s'", string(decrypted))
	}
}

func TestSetLdapConfig_NotAdmin_ReturnsForbiddden(t *testing.T) {
	env := testutil.NewTestEnv(t)

	// Create non-admin user
	email := "example@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, true, false, false, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	reqPayload := ldap.LdapConfigRequest{
		Host:          "AD01",
		Port:          389,
		BaseDN:        "dc=ez2boot,dc=org",
		BindDN:        "CN=ldap.svc,CN=Users,DC=ez2boot,DC=org",
		BindPassword:  "password",
		UseSSL:        false,
		SkipTLSVerify: false,
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("POST", "/ui/auth/ldap", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestDeleteLdapConfig_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "admin@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, true, true, false, true, "local")

	testutil.InsertLdapConfig(t, env.DB, env.Encryptor, "AD01", 389, "dc=ez2boot,dc=org", "CN=ldap.svc,CN=Users,DC=ez2boot,DC=org", "password", false, false)

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	req := httptest.NewRequest("DELETE", "/ui/auth/ldap", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var got shared.ApiResponse[any]
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !got.Success {
		t.Fatalf("want success=true, got false")
	}

	// Verify DB state - config should be gone
	var count int
	err := env.DB.QueryRow("SELECT COUNT(*) FROM ldap_config").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query ldap config: %v", err)
	}

	if count != 0 {
		t.Fatalf("want 0 ldap configs, got %d", count)
	}
}

func TestDeleteLdapConfig_NotAdmin_ReturnsForbidden(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "example@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, hash, true, false, false, true, "local")

	testutil.InsertLdapConfig(t, env.DB, env.Encryptor, "AD01", 389, "dc=ez2boot,dc=org", "CN=ldap.svc,CN=Users,DC=ez2boot,DC=org", "password", false, false)

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	req := httptest.NewRequest("DELETE", "/ui/auth/ldap", nil)
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
	err := env.DB.QueryRow("SELECT COUNT(*) FROM ldap_config").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query ldap config: %v", err)
	}

	if count != 1 {
		t.Fatalf("want 1 ldap config, got %d", count)
	}
}

// TODO mock LDAP and provide tests for remaining methods
