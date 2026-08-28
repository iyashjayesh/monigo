package monigo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"
)

func TestBasicAuthMiddleware(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create middleware
	middleware := BasicAuthMiddleware("admin", "password")
	handler := middleware(testHandler)

	// Test successful authentication
	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("admin", "password")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test failed authentication
	req = httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("admin", "wrongpassword")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAPIKeyMiddleware(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create middleware
	middleware := APIKeyMiddleware("test-api-key")
	handler := middleware(testHandler)

	// Test successful authentication via header
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", "test-api-key")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test successful authentication via query parameter
	req = httptest.NewRequest("GET", "/?api_key=test-api-key", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test failed authentication
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestIPWhitelistMiddleware(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create middleware
	middleware := IPWhitelistMiddleware([]string{"127.0.0.1", "192.168.1.0/24"})
	handler := middleware(testHandler)

	// Test allowed IP
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test blocked IP
	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "8.8.8.8:12345"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware, stop := RateLimitMiddleware(2, time.Minute)
	defer stop()
	handler := middleware(testHandler)

	// Test first request (should succeed)
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test second request (should succeed)
	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test third request (should be rate limited)
	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}
}

func TestGetSecuredUnifiedHandler(t *testing.T) {
	// Create MoniGo instance with middleware
	m := &Monigo{
		ServiceName: "test-service",
		DashboardMiddleware: []func(http.Handler) http.Handler{
			BasicAuthMiddleware("admin", "password"),
		},
	}

	// Get secured handler
	handler := GetSecuredUnifiedHandler(m)

	// Test successful authentication
	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("admin", "password")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test failed authentication
	req = httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("admin", "wrongpassword")
	w = httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestGetSecuredAPIHandlers(t *testing.T) {
	// Create MoniGo instance with middleware
	m := &Monigo{
		ServiceName: "test-service",
		APIMiddleware: []func(http.Handler) http.Handler{
			APIKeyMiddleware("test-key"),
		},
	}

	// Get secured API handlers
	handlers := GetSecuredAPIHandlers(m)

	// Test that we have handlers
	if len(handlers) == 0 {
		t.Error("Expected to have API handlers")
	}

	// Test one of the handlers
	for path, handler := range handlers {
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("X-API-Key", "test-key")
		w := httptest.NewRecorder()
		handler(w, req)

		// Should not be unauthorized (middleware should pass)
		if w.Code == http.StatusUnauthorized {
			t.Errorf("Handler for path %s should not be unauthorized", path)
		}
		break // Test just one handler
	}
}

func TestCustomAuthFunction(t *testing.T) {
	// Create MoniGo instance with custom auth function
	m := &Monigo{
		ServiceName: "test-service",
		AuthFunction: func(r *http.Request) bool {
			return r.Header.Get("X-Custom-Auth") == "valid"
		},
	}

	// Get secured handler
	handler := GetSecuredUnifiedHandler(m)

	// Test successful authentication
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Custom-Auth", "valid")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test failed authentication
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Custom-Auth", "invalid")
	w = httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// Credential comparison must be constant-time and must not accept a partial
// match. These cases previously ran through `!=`, which short-circuits on the
// first differing byte and leaks length and prefix information.
func TestBasicAuthMiddlewareRejectsPartialMatches(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := BasicAuthMiddleware("admin", "password")(testHandler)

	cases := []struct{ user, pass string }{
		{"admin", "passwor"},   // prefix of the password
		{"admin", "passwords"}, // password plus a byte
		{"admi", "password"},   // prefix of the username
		{"admins", "password"}, // username plus a byte
		{"", ""},
		{"admin", ""},
		{"", "password"},
		{"password", "admin"}, // swapped
	}

	for _, tc := range cases {
		req := httptest.NewRequest("GET", "/", nil)
		req.SetBasicAuth(tc.user, tc.pass)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("user=%q pass=%q: expected 401, got %d", tc.user, tc.pass, w.Code)
		}
	}
}

// A request carrying no credentials at all must still be challenged, so clients
// know to retry with credentials rather than treating the 401 as terminal.
func TestBasicAuthMiddlewareChallengesMissingCredentials(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := BasicAuthMiddleware("admin", "password")(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	if got := w.Header().Get("WWW-Authenticate"); got == "" {
		t.Error("expected a WWW-Authenticate challenge header on a credential-less request")
	}
}

func TestAPIKeyMiddlewareRejectsPartialMatches(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := APIKeyMiddleware("secret-key")(testHandler)

	for _, key := range []string{"", "secret-ke", "secret-keys", "s", "SECRET-KEY"} {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", key)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("key=%q: expected 401, got %d", key, w.Code)
		}
	}
}

// RateLimitMiddleware spawns a cleanup goroutine. Shutdown must stop it even
// when the caller never invokes the returned stop function.
func TestShutdownStopsRateLimiterCleanup(t *testing.T) {
	before := runtime.NumGoroutine()

	for i := 0; i < 5; i++ {
		_, _ = RateLimitMiddleware(10, 10*time.Millisecond)
	}

	m := &Monigo{ServiceName: "test-service"}
	if err := m.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}

	// Give the stopped goroutines a moment to unwind.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before+1 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("rate limiter cleanup goroutines still running after Shutdown: before=%d after=%d",
		before, runtime.NumGoroutine())
}
