package middleware

import (
	"net"
	"net/http"
	"strings"
)

func RealIP() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				r.RemoteAddr = realIP
			} else if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ips := strings.Split(forwarded, ",")
				if len(ips) > 0 {
					ip := strings.TrimSpace(ips[0])
					if net.ParseIP(ip) != nil {
						r.RemoteAddr = ip
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
