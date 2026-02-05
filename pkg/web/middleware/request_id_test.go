package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID(t *testing.T) {
	middleware := RequestID()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			t.Error("expected X-Request-ID header to be set")
		}
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(w, r)

	responseRequestID := w.Header().Get("X-Request-ID")
	if responseRequestID == "" {
		t.Error("expected X-Request-ID header in response")
	}
}

func TestRequestIDWithExistingHeader(t *testing.T) {
	middleware := RequestID()

	existingID := "existing-request-id"
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID != existingID {
			t.Errorf("expected X-Request-ID to be %s, got %s", existingID, requestID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Request-ID", existingID)

	handler.ServeHTTP(w, r)

	responseRequestID := w.Header().Get("X-Request-ID")
	if responseRequestID != existingID {
		t.Errorf("expected X-Request-ID to be %s, got %s", existingID, responseRequestID)
	}
}
