package worker_test

import (
	"context"
	"ez2boot/internal/notification"
	"ez2boot/internal/testutil"
	"ez2boot/internal/worker"
	"testing"
	"time"
)

func TestNotificationWorker_Success(t *testing.T) {
	env := testutil.NewTestEnv(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	env.Cfg.InternalClock = 1 * time.Second

	// Register stub channel
	stub := &testutil.StubNotificationChannel{}
	notification.Register(stub)

	// Create user
	adminEmail := "admin@example.com"
	adminHash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, adminEmail, &adminHash, true, true, true, true, "local")

	// Encrypt a stub config
	encConfig, err := env.Encryptor.Encrypt([]byte(`{}`))
	if err != nil {
		t.Fatalf("failed to encrypt config: %v", err)
	}

	// Insert user notification config
	_, err = env.DB.Exec("INSERT INTO user_notifications (user_id, type, config) VALUES ($1, $2, $3)", 1, "stub", encConfig)
	if err != nil {
		t.Fatalf("failed to insert user notification config: %v", err)
	}

	// Queue a notification
	_, err = env.DB.Exec("INSERT INTO notification_queue (user_id, message, title, time_added) VALUES ($1, $2, $3, $4)",
		1, "test message", "test title", time.Now().Unix())
	if err != nil {
		t.Fatalf("failed to insert notification: %v", err)
	}

	// Start notification worker
	worker.StartNotificationWorker(*env.Worker, ctx)

	// Allow time for worker to progress
	time.Sleep(1500 * time.Millisecond)

	// Verify notification was sent
	if len(stub.Calls) != 1 {
		t.Errorf("want 1 notification sent, got %d", len(stub.Calls))
	}
	if len(stub.Calls) > 0 && stub.Calls[0] != "test title" {
		t.Errorf("want title=test title, got %s", stub.Calls[0])
	}

	// Verify queue is cleared
	var count int
	err = env.DB.QueryRow("SELECT COUNT(*) FROM notification_queue").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query notification queue: %v", err)
	}
	if count != 0 {
		t.Errorf("want notification queue empty, got %d rows", count)
	}
}

// Test full lifecycle by mocking the server state and ensure system state is progressive
func TestServerSessionWorker_Success(t *testing.T) {
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
