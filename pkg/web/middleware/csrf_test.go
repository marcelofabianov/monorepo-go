package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/marcelofabianov/web/middleware"
)

func TestCSRFProtection_Disabled(t *testing.T) {
	csrf := middleware.NewCSRFProtection("secret", "csrf_token", "X-CSRF-Token", time.Hour, []string{}, false, &middleware.SecurityLogger{})

	handler := csrf.Protect()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCSRFProtection_SafeMethods(t *testing.T) {
	csrf := middleware.NewCSRFProtection("secret", "csrf_token", "X-CSRF-Token", time.Hour, []string{}, true, &middleware.SecurityLogger{})

	handler := csrf.Protect()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	methods := []string{http.MethodGet, http.MethodHead, http.MethodOptions}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200 for %s, got %d", method, w.Code)
			}
		})
	}
}

func TestCSRFProtection_ExemptRoutes(t *testing.T) {
	csrf := middleware.NewCSRFProtection("secret", "csrf_token", "X-CSRF-Token", time.Hour, []string{"/health", "/api/v1/auth/login"}, true, &middleware.SecurityLogger{})

	handler := csrf.Protect()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	exemptPaths := []string{"/health", "/api/v1/auth/login"}

	for _, path := range exemptPaths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200 for exempt path %s, got %d", path, w.Code)
			}
		})
	}
}

func TestCSRFProtection_MissingCookie(t *testing.T) {
	csrf := middleware.NewCSRFProtection("secret", "csrf_token", "X-CSRF-Token", time.Hour, []string{}, true, &middleware.SecurityLogger{})

	handler := csrf.Protect()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestCSRFProtection_MissingHeader(t *testing.T) {
	csrf := middleware.NewCSRFProtection("secret", "csrf_token", "X-CSRF-Token", time.Hour, []string{}, true, &middleware.SecurityLogger{})

	handler := csrf.Protect()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "test_token"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestCSRFProtection_InvalidToken(t *testing.T) {
	csrf := middleware.NewCSRFProtection("secret", "csrf_token", "X-CSRF-Token", time.Hour, []string{}, true, &middleware.SecurityLogger{})

	handler := csrf.Protect()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token1"})
	req.Header.Set("X-CSRF-Token", "token2")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestCSRFProtection_ValidToken(t *testing.T) {
	secLogger := &middleware.SecurityLogger{}
	csrf := middleware.NewCSRFProtection("secret", "csrf_token", "X-CSRF-Token", time.Hour, []string{}, true, secLogger)

	handler := csrf.Protect()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	token, err := csrf.GenerateToken("test-session")
	if err != nil {
		t.Fatal(err)
	}

	type contextKey string
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req = req.WithContext(context.WithValue(req.Context(), contextKey("session_id"), "test-session"))
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	req.Header.Set("X-CSRF-Token", token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCSRFProtection_GetTokenHandler(t *testing.T) {
	csrf := middleware.NewCSRFProtection("secret", "csrf_token", "X-CSRF-Token", time.Hour, []string{}, true, &middleware.SecurityLogger{})

	handler := csrf.GetTokenHandler()

	req := httptest.NewRequest(http.MethodGet, "/csrf-token", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Error("expected cookie to be set")
	}

	var found bool
	for _, cookie := range cookies {
		if cookie.Name == "csrf_token" {
			found = true
			if !cookie.HttpOnly {
				t.Error("expected HttpOnly cookie")
			}
			if !cookie.Secure {
				t.Error("expected Secure cookie")
			}
			if cookie.SameSite != http.SameSiteStrictMode {
				t.Error("expected SameSite=Strict cookie")
			}
		}
	}
	if !found {
		t.Error("csrf_token cookie not found")
	}
}

func TestCSRFProtection_GenerateToken(t *testing.T) {
	secLogger := &middleware.SecurityLogger{}
	csrf := middleware.NewCSRFProtection("secret", "csrf_token", "X-CSRF-Token", time.Hour, []string{}, true, secLogger)

	token1, err := csrf.GenerateToken("session-1")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	token2, err := csrf.GenerateToken("session-2")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if token1 == token2 {
		t.Error("tokens should be unique")
	}

	if len(token1) == 0 {
		t.Error("token should not be empty")
	}
}

func BenchmarkCSRFProtection_ValidToken(b *testing.B) {
	csrf := middleware.NewCSRFProtection("secret", "csrf_token", "X-CSRF-Token", time.Hour, []string{}, true, &middleware.SecurityLogger{})

	handler := csrf.Protect()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	token := "valid_token"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
		req.Header.Set("X-CSRF-Token", token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}
