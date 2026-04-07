package session_test

import (
	"bytes"
	"context"
	"encoding/json"
	"ez2boot/internal/session"
	"ez2boot/internal/shared"
	"ez2boot/internal/testutil"
	"ez2boot/internal/worker"
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
	testutil.InsertUser(t, env.DB, adminEmail, &adminHash, true, true, true, true, "local")

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
	req := httptest.NewRequest("POST", "/ui/session", bytes.NewReader(body))
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

	// Verify QA servers have next_state = on
	rows, err := env.DB.Query("SELECT name, next_state FROM servers WHERE server_group = $1", "QA")
	if err != nil {
		t.Fatalf("failed to query servers: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, nextState string
		if err := rows.Scan(&name, &nextState); err != nil {
			t.Fatalf("failed to scan server row: %v", err)
		}
		if nextState != "on" {
			t.Errorf("server %s: want next_state=on, got %s", name, nextState)
		}
	}
}

// Test full lifecycle by mocking the server state and ensure system state is progressive
func TestSessionLifecycle_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Set a fast clock for testing
    env.Cfg.InternalClock = 1 * time.Second

    // Start the session worker - direct call to replicate main.
	worker.StartServerSessionWorker(*env.Worker, ctx)

	adminEmail := "admin@example.com"
	adminHash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, &adminHash, true, true, true, true, "local")

	testutil.InsertServer(t, env.DB, "i-3728hvi2vn2u4vn2", "test01", "off", "QA", time.Now().Unix())
	testutil.InsertServerSession(t, env.DB, 1, "QA", time.Now().Add(2*time.Hour).Unix())

	// Server next_state is "on" here

	// Set server state on to mock behaviour
	testutil.UpdateServerState(t, env.DB, "QA", "on")

	// Allow time for worker to progress
	time.Sleep(1500 * time.Millisecond)
	
	// Verify initial notification flags
	var onNotified, toCleanup, warningNotified, offNotified int64
	err := env.DB.QueryRow(`SELECT on_notified, to_cleanup, warning_notified, off_notified FROM server_sessions 
		WHERE user_id = $1 AND server_group = $2`, 1, "QA").Scan(&onNotified, &toCleanup, &warningNotified, &offNotified)
	if err != nil {
		t.Fatalf("failed to query session flags: %v", err)
	}
	
	if onNotified != 1 {
		t.Errorf("want on_notified=1, got %d", onNotified)
	}
	if toCleanup != 0 {
		t.Errorf("want to_cleanup=0, got %d", toCleanup)
	}
	if warningNotified != 0 {
		t.Errorf("want warning_notified=0, got %d", warningNotified)
	}
	if offNotified != 0 {
		t.Errorf("want off_notified=0, got %d", offNotified)
	}

	// Allow time for worker to progress
	time.Sleep(1500 * time.Millisecond)

	// Reduce session to 10 minutes remaining - expect warning process
	testutil.UpdateServerSession(t, env.DB, "QA", time.Now().Add(10*time.Minute).Unix())

	// Allow time for worker to progress
	time.Sleep(1500 * time.Millisecond)

	// Verify warning_notified flag
	err = env.DB.QueryRow("SELECT warning_notified FROM server_sessions WHERE server_group = $1", "QA").Scan(&warningNotified)
	if err != nil {
		t.Fatalf("failed to query session flags: %v", err)
	}

	if warningNotified != 1 {
		t.Errorf("want warning_notified=1, got %d", warningNotified)
	}

	// Expire session 
	testutil.UpdateServerSession(t, env.DB, "QA", time.Now().Add(-1*time.Minute).Unix())

	// Allow time for worker to progress
	time.Sleep(1500 * time.Millisecond)

	// Expect session cleanup flag on
	err = env.DB.QueryRow("SELECT to_cleanup FROM server_sessions WHERE server_group = $1", "QA").Scan(&toCleanup)
	if err != nil {
		t.Fatalf("failed to query session flags: %v", err)
	}
	
	if toCleanup != 1 {
		t.Errorf("want to_cleanup=1, got %d", toCleanup)
	}

	// Expect next state off for servers
	rows, err := env.DB.Query("SELECT name, next_state FROM servers WHERE server_group = $1", "QA")
	if err != nil {
		t.Fatalf("failed to query servers: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, nextState string
		if err := rows.Scan(&name, &nextState); err != nil {
			t.Fatalf("failed to scan server row: %v", err)
		}
		if nextState != "off" {
			t.Errorf("server %s: want next_state=off, got %s", name, nextState)
		}
	}

	// Assume server manage worker stopped servers
	testutil.UpdateServerState(t, env.DB, "QA", "off")
	
	// Allow time for worker to progress
	time.Sleep(1500 * time.Millisecond)
	
	// Expect session to be deleted - termination and finalisation complete in same tick
	var count int
	err = env.DB.QueryRow("SELECT COUNT(*) FROM server_sessions WHERE server_group = $1", "QA").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query sessions: %v", err)
	}
	if count != 0 {
		t.Errorf("want session deleted, got %d rows", count)
	}
	
	// Expect servers next_state nil/null and state to be off
	rows, err = env.DB.Query("SELECT name, state, next_state FROM servers WHERE server_group = $1", "QA")
	if err != nil {
		t.Fatalf("failed to query servers: %v", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var name, state string
		var nextState *string
		if err := rows.Scan(&name, &state, &nextState); err != nil {
			t.Fatalf("failed to scan server row: %v", err)
		}
		if nextState != nil {
			t.Errorf("server %s: want next_state=nil, got %s", name, *nextState)
		}
	}
}

func TestAdminTerminateServerSession_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	adminHash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, &adminHash, true, true, true, true, "local")
	testutil.InsertUser(t, env.DB, "user@example.com", nil, true, false, false, true, "local")

	testutil.InsertServer(t, env.DB, "i-3728hvi2vn2u4vn2", "test01", "off", "QA", time.Now().Unix())
	testutil.InsertServerSession(t, env.DB, 2, "QA", time.Now().Add(2*time.Hour).Unix())

	cookies := testutil.LoginAndGetCookies(t, env.Router, adminEmail, adminPassword)

	reqPayload := session.ServerSessionRequest{
		ServerGroup: "QA",
		Duration:    "0h",
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("PUT", "/ui/admin/session", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify session was terminated in DB
	var expiry int64
	err := env.DB.QueryRow("SELECT expiry FROM server_sessions WHERE server_group = $1", "QA").Scan(&expiry)
	if err != nil {
		t.Fatalf("failed to query session: %v", err)
	}
	if expiry > time.Now().Unix() {
		t.Fatalf("want session terminated, got expiry in future: %d", expiry)
	}
}

func TestAdminTerminateServerSession_NonAdmin_Forbidden(t *testing.T) {
	env := testutil.NewTestEnv(t)

	adminEmail := "admin@example.com"
	adminPassword := "testpassword123"
	adminHash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, &adminHash, true, true, true, true, "local")
	testutil.InsertUser(t, env.DB, "user@example.com", &adminHash, true, false, false, true, "local")

	testutil.InsertServer(t, env.DB, "i-3728hvi2vn2u4vn2", "test01", "off", "QA", time.Now().Unix())
	testutil.InsertServerSession(t, env.DB, 2, "QA", time.Now().Add(2*time.Hour).Unix())

	cookies := testutil.LoginAndGetCookies(t, env.Router, "user@example.com", adminPassword)

	reqPayload := session.ServerSessionRequest{
		ServerGroup: "QA",
		Duration:    "0h",
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest("PUT", "/ui/admin/session", bytes.NewReader(body))
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d, body=%s", w.Code, w.Body.String())
	}

	// Verify session was not terminated
	var expiry int64
	err := env.DB.QueryRow("SELECT expiry FROM server_sessions WHERE server_group = $1", "QA").Scan(&expiry)
	if err != nil {
		t.Fatalf("failed to query session: %v", err)
	}
	if expiry < time.Now().Unix() {
		t.Fatalf("want session still active, got terminated")
	}
}
