package testutil

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"ez2boot/internal/app"
	"ez2boot/internal/auth"
	"ez2boot/internal/auth/ldap"
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/encryption"
	"ez2boot/internal/worker"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Encryptor interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

type TestEnv struct {
	DB          *sql.DB
	Logger      *slog.Logger
	Base        *db.Repository
	Cfg         *config.Config
	Router      http.Handler
	Worker      *worker.Worker
	Encryptor   Encryptor
	AuthService *auth.Service
	LdapService *ldap.Service
}

// Build test environment - in memory only
func NewTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	// Create in-memory sqlite
	testDB, err := sql.Open("sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = testDB.Close()
	})

	//logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// DB base Constructor
	baseRepo := db.NewRepository(testDB, logger)

	// Setup schema
	if err := baseRepo.SetupDB(); err != nil {
		t.Fatalf("failed to setup DB: %v", err)
	}

	cfg := &config.Config{
		CloudProvider:       "aws",
		AWSRegion:           "ap-southeast-2",
		RateLimit:           100,
		UserSessionDuration: 1 * time.Hour, // Prevent intermittent 401s during test
		EncryptionPhrase:    "newphrase",
	}

	router, services, wkr, err := app.NewApp("dev", "unknown", cfg, baseRepo, logger)
	if err != nil {
		t.Fatalf("failed to initialize app: %v", err)
	}

	encryptor, err := encryption.NewAESGCMEncryptor(cfg.EncryptionPhrase)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	return &TestEnv{
		DB:          testDB,
		Logger:      logger,
		Base:        baseRepo,
		Cfg:         cfg,
		Router:      router,
		Worker:      wkr,
		Encryptor:   encryptor,
		AuthService: services.AuthService,
		LdapService: services.LdapService,
	}
}

// Insert a dummy user into test database
func InsertUser(t *testing.T, db *sql.DB, email string, passwordHash *string, isActive bool, isAdmin bool, apiEnabled bool, uiEnabled bool, identityProvider string) {
	t.Helper()

	_, err := db.Exec(`INSERT INTO users (email, password_hash, is_active, is_admin, api_enabled, ui_enabled, identity_provider)
        				VALUES ($1, $2, $3, $4, $5, $6, $7)`, email, passwordHash, isActive, isAdmin, apiEnabled, uiEnabled, identityProvider)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
}

// Insert a dummy server session into test database
func InsertServer(t *testing.T, db *sql.DB, uniqueID string, name string, state string, serverGroup string, timeAdded int64) {
	t.Helper()

	_, err := db.Exec(`INSERT INTO servers (unique_id, name, state, server_group, time_added)
        				VALUES ($1, $2, $3, $4, $5)`, uniqueID, name, state, serverGroup, timeAdded)
	if err != nil {
		t.Fatalf("failed to insert server: %v", err)
	}
}

// Logs in a UI user and returns cookie - use in tests other than login flow
func LoginAndGetCookies(t *testing.T, router http.Handler, email, password string) []*http.Cookie {
	t.Helper()

	payload := auth.UserLogin{
		Email:    email,
		Password: password,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/ui/user/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login failed, expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("login did not set a cookie")
	}

	return cookies
}

// Inserts a pending new server session
func InsertServerSession(t *testing.T, db *sql.DB, userID int64, serverGroup string, expiry int64) {
	t.Helper()

	// Set server table for state worker
	if _, err := db.Exec("UPDATE servers SET next_state = $1, time_last_on = $2, last_user_id = $3 WHERE server_group = $4", "on", time.Now().Unix(), userID, serverGroup); err != nil {
		t.Fatal("failed to update server state")
	}

	// Set server session
	if _, err := db.Exec("INSERT INTO server_sessions (user_id, server_group, expiry, warning_notified, on_notified) VALUES ($1, $2, $3, $4, $5)", userID, serverGroup, expiry, 0, 0); err != nil {
		t.Fatal("failed to insert new server session")
	}
}

func InsertLdapConfig(t *testing.T, db *sql.DB, encryptor Encryptor, host string, port int64, baseDN string, bindDN string, bindPassword string, useSSL bool, skipTLSVerify bool) {
	t.Helper()

	// Encrypt password
	encryptedBytes, err := encryptor.Encrypt([]byte(bindPassword))
	if err != nil {
		t.Fatal("failed to encrypt password")
	}

	if _, err := db.Exec("INSERT INTO ldap_config (id, host, port, base_dn, bind_dn, bind_password, use_ssl, skip_tls_verify) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", 1, host, port, baseDN, bindDN, encryptedBytes, useSSL, skipTLSVerify); err != nil {
		t.Fatal("failed to insert ldap config")
	}

}
