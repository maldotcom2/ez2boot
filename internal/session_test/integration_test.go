package session_test

import (
	"bytes"
	"encoding/json"
	"ez2boot/internal/session"
	"ez2boot/internal/shared"
	"ez2boot/internal/testutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewServerSession_Success(t *testing.T) {
	// Create a test env
	env := testutil.NewTestEnv(t)

	// Create users
	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	adminHash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, adminHash, true, true, true, true)

	// Create servers
	testutil.InsertServer(t, env.DB, "i-3728hvi2vn2u4vn2", "test01", "off", "QA", time.Now().Unix())
	testutil.InsertServer(t, env.DB, "i-32893uhiuvuivnvj", "test02", "off", "QA", time.Now().Unix())
	testutil.InsertServer(t, env.DB, "i-3298h98h4unvunur", "test03", "off", "QA", time.Now().Unix())
	testutil.InsertServer(t, env.DB, "i-453uvbu5894uvbdu", "dev01", "off", "DEV", time.Now().Unix())

	// Login
	cookies := testutil.LoginAndGetCookies(t, env.Router, adminEmail, adminPassword)

	// Prepare HTTP request to the real route
	reqPayload := session.ServerSessionRequest{
		ServerGroup: "QA",
		Duration:    "1h",
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("POST", "/ui/session/new", bytes.NewReader(body))
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

	// Decode response
	var got shared.ApiResponse[session.ServerSessionResponse]
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	wantServerGroup := "QA"
	wantDuration := "1h"
	wantExpiry := time.Now().Add(1 * time.Hour)

	if wantServerGroup != got.Data.ServerGroup {
		t.Fatalf("server group mismatch, want: %s, got: %s", wantServerGroup, got.Data.ServerGroup)
	}

	if wantDuration != got.Data.Duration {
		t.Fatalf("duration mismatch, want: %s, got: %s", wantDuration, got.Data.Duration)
	}

	// Allow small variance
	if got.Data.Expiry.Before(wantExpiry.Add(-1*time.Second)) || got.Data.Expiry.After(wantExpiry.Add(1*time.Second)) {
		t.Fatalf("expiry mismatch, got: %v, want: ~%v", got.Data.Expiry, wantExpiry)
	}

	// Verify DB state
	var sg string
	var exp int64
	maxExp := (wantExpiry.Unix() + 1)
	minExp := (wantExpiry.Unix() - 1)
	err := env.DB.QueryRow("SELECT server_group, expiry FROM server_sessions WHERE user_id = $1", 1).Scan(&sg, &exp)
	if err != nil {
		t.Fatalf("Failed to select value: %v", err)
	}

	if exp < minExp || exp > maxExp {
		t.Fatalf("Session expiry incorrect: want range: %d-%d, got: %d", minExp, maxExp, exp)
	}

	if sg != wantServerGroup {
		t.Fatalf("Server group incorrect: want: %s, got: %s", wantServerGroup, sg)
	}
}
