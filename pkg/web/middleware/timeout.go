package middleware

import (
	"context"
	"net/http"
	"time"
)

func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})
			var panicVal interface{}

			go func() {
				defer func() {
					if p := recover(); p != nil {
						panicVal = p
					}
					close(done)
				}()

				next.ServeHTTP(w, r.WithContext(ctx))
			}()

			select {
			case <-done:
				if panicVal != nil {
					panic(panicVal)
				}
				return
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusRequestTimeout)
					_, _ = w.Write([]byte(`{"error":"request timeout"}`))
				}
				return
			}
		})
	}
}
