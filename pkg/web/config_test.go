package web_test

import (
"os"
"testing"

"github.com/marcelofabianov/web"
)

func TestLoadConfig(t *testing.T) {
origHost := os.Getenv("WEB_HTTP_HOST")
origPort := os.Getenv("WEB_HTTP_PORT")
defer func() {
os.Setenv("WEB_HTTP_HOST", origHost)
os.Setenv("WEB_HTTP_PORT", origPort)
}()

t.Run("loads defaults when no env vars set", func(t *testing.T) {
os.Unsetenv("WEB_HTTP_HOST")
os.Unsetenv("WEB_HTTP_PORT")

cfg, err := web.LoadConfig()
if err != nil {
t.Fatalf("LoadConfig() error = %v", err)
}

if cfg.HTTP.Host != "0.0.0.0" {
t.Errorf("expected host 0.0.0.0, got %s", cfg.HTTP.Host)
}
if cfg.HTTP.Port != 8080 {
t.Errorf("expected port 8080, got %d", cfg.HTTP.Port)
}
if !cfg.HTTP.CORS.Enabled {
t.Error("expected CORS to be enabled by default")
}
})

t.Run("loads from environment variables", func(t *testing.T) {
os.Setenv("WEB_HTTP_HOST", "localhost")
os.Setenv("WEB_HTTP_PORT", "3000")
os.Setenv("WEB_HTTP_CORS_ENABLED", "false")
defer os.Unsetenv("WEB_HTTP_CORS_ENABLED")

cfg, err := web.LoadConfig()
if err != nil {
t.Fatalf("LoadConfig() error = %v", err)
}

if cfg.HTTP.Host != "localhost" {
t.Errorf("expected host localhost, got %s", cfg.HTTP.Host)
}
if cfg.HTTP.Port != 3000 {
t.Errorf("expected port 3000, got %d", cfg.HTTP.Port)
}
if cfg.HTTP.CORS.Enabled {
t.Error("expected CORS to be disabled")
}
})
}
