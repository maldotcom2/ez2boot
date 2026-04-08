package middleware_test

import (
	"ez2boot/internal/testutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimit_PublicEndpointBlocked(t *testing.T) {
	env := testutil.NewTestEnv(t)

	// Exhaust the bucket
	for i := 0; i < env.Cfg.PublicRateLimit*2+1; i++ {
		req := httptest.NewRequest("POST", "/ui/auth/login", nil)
		w := httptest.NewRecorder()
		env.Router.ServeHTTP(w, req)

		if w.Code == http.StatusTooManyRequests {
			return // Got blocked as expected
		}
	}

	t.Fatal("want 429 after burst exceeded, got none")
}

func TestRateLimit_PrivateEndpointBlocked(t *testing.T) {
	env := testutil.NewTestEnv(t)

	email := "example@example.com"
	password := "testpassword123"
	hash := "$argon2id$v=19$m=131072,t=4,p=1$bBVby41uAKJ7KghSdCEt8g$80aCufSfLP2tAZ9bxAjbs8mArxgjmgrP3UkPn8MKCJY"
	testutil.InsertUser(t, env.DB, email, &hash, true, true, false, true, "local")

	cookies := testutil.LoginAndGetCookies(t, env.Router, email, password)

	// Exhaust the bucket, +10 to for headroom
	for i := 0; i < env.Cfg.PrivateRateLimit*2+10; i++ {
		req := httptest.NewRequest("GET", "/ui/users", nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		w := httptest.NewRecorder()
		env.Router.ServeHTTP(w, req)

		if w.Code == http.StatusTooManyRequests {
			return
		}
	}

	t.Fatal("want 429 after burst exceeded, got none")
}

func TestRateLimit_PublicEndpointRecovery(t *testing.T) {
	env := testutil.NewTestEnv(t)

	// Exhaust the bucket
	for i := 0; i < env.Cfg.PublicRateLimit*2+1; i++ {
		req := httptest.NewRequest("POST", "/ui/auth/login", nil)
		w := httptest.NewRecorder()
		env.Router.ServeHTTP(w, req)
	}

	// Wait for bucket to refill one token
	time.Sleep(time.Second / time.Duration(env.Cfg.PublicRateLimit))

	// Should be allowed again
	req := httptest.NewRequest("POST", "/ui/auth/login", nil)
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code == http.StatusTooManyRequests {
		t.Fatal("want request allowed after refill, got 429")
	}
}
