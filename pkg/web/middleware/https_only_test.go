package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPSOnly_HTTPRequest(t *testing.T) {
	// Arrange
	cfg := HTTPSOnlyConfig{Enabled: true}
	handler := HTTPSOnly(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "example.com"
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

	body := rec.Body.String()
	assert.Contains(t, body, "HTTPS_REQUIRED")
	assert.Contains(t, body, "This server only accepts HTTPS connections")
	assert.Contains(t, body, "https://example.com/test")
}

func TestHTTPSOnly_HTTPSRequest(t *testing.T) {
	// Arrange
	cfg := HTTPSOnlyConfig{Enabled: true}
	handler := HTTPSOnly(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "example.com"
	req.TLS = &tls.ConnectionState{} // Simulate TLS connection
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())
}

func TestHTTPSOnly_Disabled(t *testing.T) {
	// Arrange
	cfg := HTTPSOnlyConfig{Enabled: false}
	handler := HTTPSOnly(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "example.com"
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert - should pass through even with HTTP
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())
}

func TestHTTPSOnly_CustomRedirectURL(t *testing.T) {
	// Arrange
	cfg := HTTPSOnlyConfig{
		Enabled:     true,
		RedirectURL: "https://custom.example.com/api",
	}
	handler := HTTPSOnly(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "example.com"
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	require.Equal(t, http.StatusBadRequest, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "https://custom.example.com/api")
	assert.NotContains(t, body, "https://example.com/test")
}

func TestHTTPSOnly_PreservesQueryParams(t *testing.T) {
	// Arrange
	cfg := HTTPSOnlyConfig{Enabled: true}
	handler := HTTPSOnly(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users?page=2&limit=10", nil)
	req.Host = "example.com"
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	require.Equal(t, http.StatusBadRequest, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "https://example.com/api/users?page=2&limit=10")
}

func TestHTTPSOnly_PreservesHostAndPath(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		path        string
		expectedURL string
	}{
		{
			name:        "root path",
			host:        "api.example.com",
			path:        "/",
			expectedURL: "https://api.example.com/",
		},
		{
			name:        "nested path",
			host:        "api.example.com",
			path:        "/v1/users/123",
			expectedURL: "https://api.example.com/v1/users/123",
		},
		{
			name:        "with port",
			host:        "localhost:8080",
			path:        "/health",
			expectedURL: "https://localhost:8080/health",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg := HTTPSOnlyConfig{Enabled: true}
			handler := HTTPSOnly(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Host = tt.host
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			require.Equal(t, http.StatusBadRequest, rec.Code)
			body := rec.Body.String()
			assert.Contains(t, body, tt.expectedURL)
		})
	}
}

func TestHTTPSOnly_JSONResponse(t *testing.T) {
	// Arrange
	cfg := HTTPSOnlyConfig{Enabled: true}
	handler := HTTPSOnly(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "example.com"
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	body := rec.Body.String()

	// Check JSON structure
	assert.True(t, strings.HasPrefix(body, "{"))
	assert.True(t, strings.HasSuffix(strings.TrimSpace(body), "}"))
	assert.Contains(t, body, `"error"`)
	assert.Contains(t, body, `"code"`)
	assert.Contains(t, body, `"message"`)
	assert.Contains(t, body, `"details"`)
	assert.Contains(t, body, `"https_url"`)
}
