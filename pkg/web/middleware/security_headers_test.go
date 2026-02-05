package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	cfg := SecurityHeadersConfig{
		XContentTypeOptions:     "nosniff",
		XFrameOptions:           "DENY",
		ContentSecurityPolicy:   "default-src 'none'",
		ReferrerPolicy:          "no-referrer",
		StrictTransportSecurity: "max-age=31536000",
		CacheControl:            "no-store",
		PermissionsPolicy:       "camera=()",
		XDNSPrefetchControl:     "off",
		XDownloadOptions:        "noopen",
	}

	middleware := SecurityHeaders(cfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(w, r)

	tests := []struct {
		header   string
		expected string
	}{
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"Content-Security-Policy", "default-src 'none'"},
		{"Referrer-Policy", "no-referrer"},
		{"Strict-Transport-Security", "max-age=31536000"},
		{"Cache-Control", "no-store"},
		{"Permissions-Policy", "camera=()"},
		{"X-DNS-Prefetch-Control", "off"},
		{"X-Download-Options", "noopen"},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			value := w.Header().Get(tt.header)
			if value != tt.expected {
				t.Errorf("expected %s to be %s, got %s", tt.header, tt.expected, value)
			}
		})
	}
}
