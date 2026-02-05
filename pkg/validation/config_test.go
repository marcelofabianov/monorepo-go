package validation_test

import (
"os"
"testing"

"github.com/marcelofabianov/validation"
)

func TestLoadConfig(t *testing.T) {
origLogging := os.Getenv("VALIDATION_ENABLE_LOGGING")
origSanitize := os.Getenv("VALIDATION_SANITIZE_SENSITIVE_DATA")
defer func() {
os.Setenv("VALIDATION_ENABLE_LOGGING", origLogging)
os.Setenv("VALIDATION_SANITIZE_SENSITIVE_DATA", origSanitize)
}()

t.Run("loads defaults when no env vars set", func(t *testing.T) {
os.Unsetenv("VALIDATION_ENABLE_LOGGING")
os.Unsetenv("VALIDATION_SANITIZE_SENSITIVE_DATA")

cfg, err := validation.LoadConfig()
if err != nil {
t.Fatalf("LoadConfig() error = %v", err)
}

if !cfg.EnableLogging {
t.Error("expected enable logging to be true")
}
if !cfg.SanitizeSensitiveData {
t.Error("expected sanitize sensitive data to be true")
}
if cfg.LogSuccessfulValidations {
t.Error("expected log successful validations to be false")
}
})

t.Run("loads from environment variables", func(t *testing.T) {
os.Setenv("VALIDATION_ENABLE_LOGGING", "false")
os.Setenv("VALIDATION_SANITIZE_SENSITIVE_DATA", "false")
os.Setenv("VALIDATION_LOG_SUCCESSFUL_VALIDATIONS", "true")
defer func() {
os.Unsetenv("VALIDATION_LOG_SUCCESSFUL_VALIDATIONS")
}()

cfg, err := validation.LoadConfig()
if err != nil {
t.Fatalf("LoadConfig() error = %v", err)
}

if cfg.EnableLogging {
t.Error("expected enable logging to be false")
}
if cfg.SanitizeSensitiveData {
t.Error("expected sanitize sensitive data to be false")
}
if !cfg.LogSuccessfulValidations {
t.Error("expected log successful validations to be true")
}
})
}

func TestDefaultConfig(t *testing.T) {
cfg := validation.DefaultConfig()

if !cfg.EnableLogging {
t.Error("expected enable logging to be true")
}
if !cfg.SanitizeSensitiveData {
t.Error("expected sanitize sensitive data to be true")
}
if cfg.LogSuccessfulValidations {
t.Error("expected log successful validations to be false")
}
if len(cfg.AdditionalSensitiveFields) != 0 {
t.Errorf("expected empty additional sensitive fields, got %d", len(cfg.AdditionalSensitiveFields))
}
}
