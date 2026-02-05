package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAcceptJSON(t *testing.T) {
	middleware := AcceptJSON()

	tests := []struct {
		name           string
		acceptHeader   string
		expectedStatus int
	}{
		{"empty accept header", "", http.StatusOK},
		{"application/json", "application/json", http.StatusOK},
		{"application/*", "application/*", http.StatusOK},
		{"*/*", "*/*", http.StatusOK},
		{"text/html", "text/html", http.StatusNotAcceptable},
		{"text/plain", "text/plain", http.StatusNotAcceptable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.acceptHeader != "" {
				r.Header.Set("Accept", tt.acceptHeader)
			}

			handler.ServeHTTP(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
