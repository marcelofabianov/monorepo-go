package middleware

import (
	"net/http"
)

func SecurityHeaders(cfg SecurityHeadersConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.XContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", cfg.XContentTypeOptions)
			}
			if cfg.XFrameOptions != "" {
				w.Header().Set("X-Frame-Options", cfg.XFrameOptions)
			}
			if cfg.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
			}
			if cfg.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
			}
			if cfg.StrictTransportSecurity != "" {
				w.Header().Set("Strict-Transport-Security", cfg.StrictTransportSecurity)
			}
			if cfg.CacheControl != "" {
				w.Header().Set("Cache-Control", cfg.CacheControl)
			}
			if cfg.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", cfg.PermissionsPolicy)
			}
			if cfg.XDNSPrefetchControl != "" {
				w.Header().Set("X-DNS-Prefetch-Control", cfg.XDNSPrefetchControl)
			}
			if cfg.XDownloadOptions != "" {
				w.Header().Set("X-Download-Options", cfg.XDownloadOptions)
			}

			next.ServeHTTP(w, r)
		})
	}
}
