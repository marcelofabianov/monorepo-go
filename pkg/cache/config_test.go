package cache_test

import (
"os"
"testing"
"time"

"github.com/marcelofabianov/cache"
)

func TestLoadConfig(t *testing.T) {
origHost := os.Getenv("CACHE_REDIS_HOST")
origPort := os.Getenv("CACHE_REDIS_PORT")
defer func() {
os.Setenv("CACHE_REDIS_HOST", origHost)
os.Setenv("CACHE_REDIS_PORT", origPort)
}()

t.Run("loads defaults when no env vars set", func(t *testing.T) {
os.Unsetenv("CACHE_REDIS_HOST")
os.Unsetenv("CACHE_REDIS_PORT")

cfg, err := cache.LoadConfig()
if err != nil {
t.Fatalf("LoadConfig() error = %v", err)
}

if cfg.Redis.Credentials.Host != "localhost" {
t.Errorf("expected host localhost, got %s", cfg.Redis.Credentials.Host)
}
if cfg.Redis.Credentials.Port != 6379 {
t.Errorf("expected port 6379, got %d", cfg.Redis.Credentials.Port)
}
if cfg.Redis.Pool.MaxIdleConns != 10 {
t.Errorf("expected max idle conns 10, got %d", cfg.Redis.Pool.MaxIdleConns)
}
if cfg.Redis.Pool.MaxActiveConns != 20 {
t.Errorf("expected max active conns 20, got %d", cfg.Redis.Pool.MaxActiveConns)
}
})

t.Run("loads from environment variables", func(t *testing.T) {
os.Setenv("CACHE_REDIS_HOST", "redis-server")
os.Setenv("CACHE_REDIS_PORT", "6380")
os.Setenv("CACHE_REDIS_PASSWORD", "secret")
os.Setenv("CACHE_REDIS_DB", "1")
defer func() {
os.Unsetenv("CACHE_REDIS_PASSWORD")
os.Unsetenv("CACHE_REDIS_DB")
}()

cfg, err := cache.LoadConfig()
if err != nil {
t.Fatalf("LoadConfig() error = %v", err)
}

if cfg.Redis.Credentials.Host != "redis-server" {
t.Errorf("expected host redis-server, got %s", cfg.Redis.Credentials.Host)
}
if cfg.Redis.Credentials.Port != 6380 {
t.Errorf("expected port 6380, got %d", cfg.Redis.Credentials.Port)
}
if cfg.Redis.Credentials.Password != "secret" {
t.Errorf("expected password secret, got %s", cfg.Redis.Credentials.Password)
}
if cfg.Redis.Credentials.DB != 1 {
t.Errorf("expected db 1, got %d", cfg.Redis.Credentials.DB)
}
})

t.Run("validates invalid port", func(t *testing.T) {
os.Setenv("CACHE_REDIS_PORT", "99999")
defer os.Unsetenv("CACHE_REDIS_PORT")

_, err := cache.LoadConfig()
if err == nil {
t.Error("expected error for invalid port")
}
})
}

func TestConfigProvider(t *testing.T) {
cfg := &cache.Config{
Redis: cache.RedisConfig{
Credentials: cache.RedisCredentialsConfig{
Host:     "localhost",
Port:     6379,
Password: "pass",
DB:       0,
},
Pool: cache.RedisPoolConfig{
MaxIdleConns:   5,
MaxActiveConns: 10,
},
Connect: cache.RedisConnectConfig{
QueryTimeout:   2 * time.Second,
ExecTimeout:    3 * time.Second,
BackoffMin:     100 * time.Millisecond,
BackoffMax:     5 * time.Second,
BackoffFactor:  2,
BackoffJitter:  true,
BackoffRetries: 3,
},
},
}

var _ cache.ConfigProvider = cfg

if cfg.GetHost() != "localhost" {
t.Errorf("GetHost() = %s, want localhost", cfg.GetHost())
}
if cfg.GetPort() != 6379 {
t.Errorf("GetPort() = %d, want 6379", cfg.GetPort())
}
if cfg.GetPassword() != "pass" {
t.Errorf("GetPassword() = %s, want pass", cfg.GetPassword())
}
if cfg.GetDB() != 0 {
t.Errorf("GetDB() = %d, want 0", cfg.GetDB())
}
if cfg.GetMaxIdleConns() != 5 {
t.Errorf("GetMaxIdleConns() = %d, want 5", cfg.GetMaxIdleConns())
}
if cfg.GetMaxActiveConns() != 10 {
t.Errorf("GetMaxActiveConns() = %d, want 10", cfg.GetMaxActiveConns())
}
if cfg.GetQueryTimeout() != 2*time.Second {
t.Errorf("GetQueryTimeout() = %v, want 2s", cfg.GetQueryTimeout())
}
if cfg.GetExecTimeout() != 3*time.Second {
t.Errorf("GetExecTimeout() = %v, want 3s", cfg.GetExecTimeout())
}
if cfg.GetBackoffMin() != 100*time.Millisecond {
t.Errorf("GetBackoffMin() = %v, want 100ms", cfg.GetBackoffMin())
}
if cfg.GetBackoffMax() != 5*time.Second {
t.Errorf("GetBackoffMax() = %v, want 5s", cfg.GetBackoffMax())
}
if cfg.GetBackoffFactor() != 2 {
t.Errorf("GetBackoffFactor() = %d, want 2", cfg.GetBackoffFactor())
}
if cfg.GetBackoffJitter() != true {
t.Error("GetBackoffJitter() = false, want true")
}
if cfg.GetBackoffRetries() != 3 {
t.Errorf("GetBackoffRetries() = %d, want 3", cfg.GetBackoffRetries())
}
}

func TestGetRedisRetryConfig(t *testing.T) {
cfg := &cache.Config{
Redis: cache.RedisConfig{
Connect: cache.RedisConnectConfig{
BackoffMin:     100 * time.Millisecond,
BackoffMax:     5 * time.Second,
BackoffFactor:  2,
BackoffJitter:  true,
BackoffRetries: 5,
},
},
}

retryConfig := cfg.GetRedisRetryConfig()
if retryConfig == nil {
t.Fatal("GetRedisRetryConfig() should not return nil")
}

if retryConfig.MaxAttempts != 5 {
t.Errorf("MaxAttempts = %v, want 5", retryConfig.MaxAttempts)
}

if retryConfig.Strategy == nil {
t.Error("Strategy should not be nil")
}
}
