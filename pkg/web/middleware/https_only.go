package middleware

import (
	"fmt"
	"net/http"
)

// HTTPSOnlyConfig holds configuration for HTTPS enforcement
type HTTPSOnlyConfig struct {
	Enabled     bool
	RedirectURL string
}

// HTTPSOnly is a middleware that ensures all requests are made over HTTPS.
// If a request is made over HTTP (detected by TLS being nil), it returns a
// 400 Bad Request with a message instructing the client to use HTTPS.
//
// This middleware is particularly useful when running a server that only
// accepts HTTPS connections but may receive HTTP requests from misconfigured
// clients.
//
// Example usage:
//
//	config := middleware.HTTPSOnlyConfig{
//		Enabled:     true,
//		RedirectURL: "", // optional custom URL
//	}
//	r := chi.NewRouter()
//	r.Use(middleware.HTTPSOnly(config))
func HTTPSOnly(cfg HTTPSOnlyConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip middleware if not enabled
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Check if the connection is using TLS
			if r.TLS == nil {
				// Connection is not encrypted (HTTP)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusBadRequest)

				// Use custom redirect URL or build from request
				httpsURL := cfg.RedirectURL
				if httpsURL == "" {
					httpsURL = fmt.Sprintf("https://%s%s", r.Host, r.RequestURI)
				}

				response := fmt.Sprintf(`{
  "error": {
    "code": "HTTPS_REQUIRED",
    "message": "This server only accepts HTTPS connections",
    "details": "Please use HTTPS instead of HTTP",
    "https_url": "%s"
  }
}`, httpsURL)

				_, _ = w.Write([]byte(response))
				return
			}

			// Connection is encrypted (HTTPS), continue
			next.ServeHTTP(w, r)
		})
	}
}
