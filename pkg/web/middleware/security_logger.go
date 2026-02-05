package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type SecurityEventType string

const (
	EventCSRFViolation      SecurityEventType = "csrf_violation"
	EventRateLimitExceeded  SecurityEventType = "rate_limit_exceeded"
	EventInvalidAuth        SecurityEventType = "invalid_auth"
	EventSuspiciousActivity SecurityEventType = "suspicious_activity"
	EventIPSpoofing         SecurityEventType = "ip_spoofing"
	EventLoginSuccess       SecurityEventType = "login_success"
	EventLoginFailed        SecurityEventType = "login_failed"
	EventAccountLocked      SecurityEventType = "account_locked"
	EventPasswordChanged    SecurityEventType = "password_changed"
	EventTokenRefreshed     SecurityEventType = "token_refreshed"
	EventTokenRevoked       SecurityEventType = "token_revoked"
)

type SecuritySeverity string

const (
	SeverityLow      SecuritySeverity = "low"
	SeverityMedium   SecuritySeverity = "medium"
	SeverityHigh     SecuritySeverity = "high"
	SeverityCritical SecuritySeverity = "critical"
)

type SecurityLogger struct {
	logger *slog.Logger
}

func NewSecurityLogger(log *slog.Logger) *SecurityLogger {
	return &SecurityLogger{logger: log}
}

func (s *SecurityLogger) LogEvent(eventType SecurityEventType, severity SecuritySeverity, r *http.Request, details map[string]string) {
	if s == nil || s.logger == nil {
		return
	}

	args := []any{
		"event_type", string(eventType),
		"severity", string(severity),
		"ip", getRealIP(r),
		"path", r.URL.Path,
		"method", r.Method,
		"user_agent", r.UserAgent(),
		"timestamp", time.Now().UTC().Format(time.RFC3339),
	}

	for k, v := range details {
		args = append(args, k, v)
	}

	switch severity {
	case SeverityCritical, SeverityHigh:
		s.logger.Error("security_event", args...)
	case SeverityMedium:
		s.logger.Warn("security_event", args...)
	default:
		s.logger.Info("security_event", args...)
	}
}

func (s *SecurityLogger) LogCSRFViolation(r *http.Request, reason string) {
	s.LogEvent(EventCSRFViolation, SeverityHigh, r, map[string]string{
		"reason": reason,
	})
}

func (s *SecurityLogger) LogRateLimitExceeded(r *http.Request, limit int, window string) {
	s.LogEvent(EventRateLimitExceeded, SeverityMedium, r, map[string]string{
		"limit":  string(rune(limit)),
		"window": window,
	})
}

func (s *SecurityLogger) LogIPSpoofing(r *http.Request, suspectedIP string) {
	s.LogEvent(EventIPSpoofing, SeverityCritical, r, map[string]string{
		"suspected_ip": suspectedIP,
	})
}

func (s *SecurityLogger) LogAuthEvent(eventType SecurityEventType, email string, r *http.Request, success bool, reason string) {
	if s == nil || s.logger == nil {
		return
	}

	severity := SeverityMedium
	switch eventType {
	case EventLoginFailed, EventAccountLocked:
		severity = SeverityHigh
	case EventLoginSuccess, EventTokenRefreshed, EventTokenRevoked, EventPasswordChanged:
		severity = SeverityLow
	}

	details := map[string]string{
		"email":   email,
		"success": boolToString(success),
	}

	if reason != "" {
		details["reason"] = reason
	}

	s.LogEvent(eventType, severity, r, details)
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func getRealIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}
