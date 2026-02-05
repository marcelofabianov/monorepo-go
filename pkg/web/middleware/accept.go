package middleware

import (
	"net/http"
	"strings"
)

func AcceptJSON() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accept := r.Header.Get("Accept")

			if accept == "" {
				next.ServeHTTP(w, r)
				return
			}

			if acceptsJSON(accept) {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusNotAcceptable)
			_, _ = w.Write([]byte(`{"code":"NOT_ACCEPTABLE","message":"This API only returns application/json","status_code":406}`))
		})
	}
}

func acceptsJSON(accept string) bool {
	accept = strings.ToLower(accept)

	acceptableTypes := []string{
		"application/json",
		"application/*",
		"*/*",
	}

	for _, acceptable := range acceptableTypes {
		if strings.Contains(accept, acceptable) {
			return true
		}
	}

	return false
}
