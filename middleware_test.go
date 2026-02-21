package monigo

import (
	"net/http"
	"net/http/httptest"
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
