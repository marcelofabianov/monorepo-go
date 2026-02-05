package middleware_test

import (
"log/slog"
"net/http"
"net/http/httptest"
"os"
"testing"

"github.com/marcelofabianov/web/middleware"
"github.com/stretchr/testify/assert"
)

func setupSecurityLogger(t *testing.T) *middleware.SecurityLogger {
t.Helper()
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
Level: slog.LevelInfo,
}))
return middleware.NewSecurityLogger(logger)
}

func TestSecurityLogger_LogAuthEvent_LoginSuccess(t *testing.T) {
sl := setupSecurityLogger(t)
req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
req.Header.Set("User-Agent", "Test-Agent")
req.RemoteAddr = "192.168.1.1:12345"

sl.LogAuthEvent(
middleware.EventLoginSuccess,
"user@example.com",
req,
true,
"",
)
}

func TestSecurityLogger_LogAuthEvent_LoginFailed(t *testing.T) {
sl := setupSecurityLogger(t)
req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)

sl.LogAuthEvent(
middleware.EventLoginFailed,
"user@example.com",
req,
false,
"invalid_password",
)
}

func TestSecurityLogger_LogAuthEvent_AccountLocked(t *testing.T) {
sl := setupSecurityLogger(t)
req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)

sl.LogAuthEvent(
middleware.EventAccountLocked,
"user@example.com",
req,
false,
"max_attempts_exceeded",
)
}

func TestSecurityLogger_LogAuthEvent_TokenRefreshed(t *testing.T) {
sl := setupSecurityLogger(t)
req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)

sl.LogAuthEvent(
middleware.EventTokenRefreshed,
"user@example.com",
req,
true,
"",
)
}

func TestSecurityLogger_LogAuthEvent_TokenRevoked(t *testing.T) {
sl := setupSecurityLogger(t)
req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)

sl.LogAuthEvent(
middleware.EventTokenRevoked,
"user@example.com",
req,
true,
"user_logout",
)
}

func TestSecurityLogger_LogAuthEvent_PasswordChanged(t *testing.T) {
sl := setupSecurityLogger(t)
req := httptest.NewRequest(http.MethodPut, "/users/me/password", nil)

sl.LogAuthEvent(
middleware.EventPasswordChanged,
"user@example.com",
req,
true,
"",
)
}

func TestSecurityLogger_LogAuthEvent_NilLogger(t *testing.T) {
sl := middleware.NewSecurityLogger(nil)
req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)

assert.NotPanics(t, func() {
sl.LogAuthEvent(
middleware.EventLoginSuccess,
"user@example.com",
req,
true,
"",
)
})
}
