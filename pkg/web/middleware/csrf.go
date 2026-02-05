package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CSRFProtection struct {
	secret         []byte
	cookieName     string
	headerName     string
	ttl            time.Duration
	exemptPaths    map[string]bool
	enabled        bool
	securityLogger *SecurityLogger
}

func NewCSRFProtection(secret, cookieName, headerName string, ttl time.Duration, exempt []string, enabled bool, secLogger *SecurityLogger) *CSRFProtection {
	secretBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil || len(secretBytes) < 32 {
		secretBytes = []byte(secret)
	}

	exemptMap := make(map[string]bool, len(exempt))
	for _, path := range exempt {
		exemptMap[path] = true
	}

	return &CSRFProtection{
		secret:         secretBytes,
		cookieName:     cookieName,
		headerName:     headerName,
		ttl:            ttl,
		exemptPaths:    exemptMap,
		enabled:        enabled,
		securityLogger: secLogger,
	}
}

func (c *CSRFProtection) Protect() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !c.enabled || c.isSafeMethod(r.Method) || c.isExempt(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(c.cookieName)
			if err != nil {
				if c.securityLogger != nil {
					c.securityLogger.LogCSRFViolation(r, "cookie_missing")
				}
				http.Error(w, "CSRF token missing", http.StatusForbidden)
				return
			}

			headerToken := r.Header.Get(c.headerName)
			if headerToken == "" {
				if c.securityLogger != nil {
					c.securityLogger.LogCSRFViolation(r, "header_missing")
				}
				http.Error(w, "CSRF token missing in header", http.StatusForbidden)
				return
			}

			sessionID := c.getSessionID(r)
			if !c.validateToken(sessionID, cookie.Value, headerToken) {
				if c.securityLogger != nil {
					c.securityLogger.LogCSRFViolation(r, "token_invalid")
				}
				http.Error(w, "CSRF token invalid", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (c *CSRFProtection) getSessionID(r *http.Request) string {
	if sessionID := r.Context().Value("session_id"); sessionID != nil {
		return sessionID.(string)
	}
	return getRealIP(r)
}

func (c *CSRFProtection) GenerateToken(sessionID string) (string, error) {
	timestamp := time.Now().Unix()
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}

	h := hmac.New(sha256.New, c.secret)
	h.Write([]byte(sessionID))
	h.Write([]byte(strconv.FormatInt(timestamp, 10)))
	h.Write(random)

	tokenBytes := h.Sum(nil)
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	return strconv.FormatInt(timestamp, 10) + ":" + token, nil
}

func (c *CSRFProtection) SetTokenCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     c.cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(c.ttl.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}

func (c *CSRFProtection) GetTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := c.getSessionID(r)
		token, err := c.GenerateToken(sessionID)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		c.SetTokenCookie(w, token)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
			if c.securityLogger != nil {
				c.securityLogger.LogCSRFViolation(r, "response_encoding_error")
			}
		}
	}
}

func (c *CSRFProtection) validateToken(sessionID, cookieToken, headerToken string) bool {
	if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) != 1 {
		return false
	}

	parts := strings.Split(cookieToken, ":")
	if len(parts) != 2 {
		return false
	}

	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false
	}

	if time.Since(time.Unix(timestamp, 0)) > c.ttl {
		return false
	}

	expectedToken, err := c.GenerateToken(sessionID)
	if err != nil {
		return false
	}

	return strings.HasPrefix(expectedToken, parts[0]+":")
}

func (c *CSRFProtection) isSafeMethod(method string) bool {
	return method == http.MethodGet ||
		method == http.MethodHead ||
		method == http.MethodOptions
}

func (c *CSRFProtection) isExempt(path string) bool {
	return c.exemptPaths[path]
}
