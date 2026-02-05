package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	"github.com/marcelofabianov/web/middleware"
)

func TestRateLimiter_Disabled(t *testing.T) {
	secLogger := &middleware.SecurityLogger{}
	limiter := middleware.NewRateLimiter(nil, false, []string{}, secLogger)

	handler := limiter.GlobalLimit(10, time.Minute, 15)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRateLimiter_NoRedis(t *testing.T) {
	secLogger := &middleware.SecurityLogger{}
	limiter := middleware.NewRateLimiter(nil, true, []string{}, secLogger)

	handler := limiter.GlobalLimit(10, time.Minute, 15)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 (fail open), got %d", w.Code)
	}
}

func TestByIPStrategy(t *testing.T) {
	client := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: client.Addr()})
	secLogger := &middleware.SecurityLogger{}

	limiter := middleware.NewRateLimiter(redisClient, true, []string{"10.0.0.0/8"}, secLogger)
	strategy := middleware.ByIP(limiter)

	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "X-Forwarded-For present (trusted proxy)",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.1"},
			remoteAddr: "10.0.0.1:1234",
			expected:   "192.168.1.1",
		},
		{
			name:       "fallback to RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "10.0.0.1:1234",
			expected:   "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			result := strategy(req)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestByRouteStrategy(t *testing.T) {
	t.Skip("Strategy tests need refactoring after security improvements")
	// TODO: Implement new strategy tests
}

func TestCompositeStrategy(t *testing.T) {
	t.Skip("Strategy tests need refactoring after security improvements")
	// TODO: Implement new strategy tests
}

func BenchmarkRateLimiter(b *testing.B) {
	opt, _ := redis.ParseURL("redis://localhost:6379/0")
	client := redis.NewClient(opt)
	defer client.Close()

	secLogger := &middleware.SecurityLogger{}
	limiter := middleware.NewRateLimiter(client, true, []string{}, secLogger)
	handler := limiter.GlobalLimit(1000, time.Second, 1500)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler { return handler })
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}
