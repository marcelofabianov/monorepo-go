package cache

import (
"fmt"
"os"
"path/filepath"
"strings"
"time"

"github.com/marcelofabianov/retry"
"github.com/spf13/viper"
)

type ConfigProvider interface {
GetHost() string
GetPort() int
GetPassword() string
GetDB() int
GetMaxIdleConns() int
GetMaxActiveConns() int
GetQueryTimeout() time.Duration
GetExecTimeout() time.Duration
GetBackoffMin() time.Duration
GetBackoffMax() time.Duration
GetBackoffFactor() int
GetBackoffJitter() bool
GetBackoffRetries() int
}

type Config struct {
Redis RedisConfig
}

type RedisConfig struct {
Connect     RedisConnectConfig
Pool        RedisPoolConfig
Credentials RedisCredentialsConfig
}

type RedisConnectConfig struct {
QueryTimeout   time.Duration
ExecTimeout    time.Duration
BackoffMin     time.Duration
BackoffMax     time.Duration
BackoffFactor  int
BackoffJitter  bool
BackoffRetries int
}

type RedisPoolConfig struct {
MaxIdleConns   int
MaxActiveConns int
}

type RedisCredentialsConfig struct {
Host     string
Port     int
Password string
DB       int
}

var _ ConfigProvider = (*Config)(nil)

func LoadConfig() (*Config, error) {
v := viper.New()
v.SetEnvPrefix("CACHE")
v.AutomaticEnv()
v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

if envFile := findEnvFile(); envFile != "" {
v.SetConfigFile(envFile)
_ = v.ReadInConfig()
}

setDefaults(v)

cfg := &Config{
Redis: RedisConfig{
Credentials: RedisCredentialsConfig{
Host:     v.GetString("redis.host"),
Port:     v.GetInt("redis.port"),
Password: v.GetString("redis.password"),
DB:       v.GetInt("redis.db"),
},
Connect: RedisConnectConfig{
QueryTimeout:   v.GetDuration("redis.connect.query_timeout"),
ExecTimeout:    v.GetDuration("redis.connect.exec_timeout"),
BackoffMin:     v.GetDuration("redis.connect.backoff_min"),
BackoffMax:     v.GetDuration("redis.connect.backoff_max"),
BackoffFactor:  v.GetInt("redis.connect.backoff_factor"),
BackoffJitter:  v.GetBool("redis.connect.backoff_jitter"),
BackoffRetries: v.GetInt("redis.connect.backoff_retries"),
},
Pool: RedisPoolConfig{
MaxIdleConns:   v.GetInt("redis.pool.max_idle_conns"),
MaxActiveConns: v.GetInt("redis.pool.max_active_conns"),
},
},
}

if err := validateConfig(cfg); err != nil {
return nil, err
}

return cfg, nil
}

func setDefaults(v *viper.Viper) {
v.SetDefault("redis.host", "localhost")
v.SetDefault("redis.port", 6379)
v.SetDefault("redis.password", "")
v.SetDefault("redis.db", 0)
v.SetDefault("redis.connect.query_timeout", 2*time.Second)
v.SetDefault("redis.connect.exec_timeout", 2*time.Second)
v.SetDefault("redis.connect.backoff_min", 200*time.Millisecond)
v.SetDefault("redis.connect.backoff_max", 15*time.Second)
v.SetDefault("redis.connect.backoff_factor", 2)
v.SetDefault("redis.connect.backoff_jitter", true)
v.SetDefault("redis.connect.backoff_retries", 7)
v.SetDefault("redis.pool.max_idle_conns", 10)
v.SetDefault("redis.pool.max_active_conns", 20)
}

func findEnvFile() string {
dir, err := os.Getwd()
if err != nil {
return ""
}

for i := 0; i < 5; i++ {
envPath := filepath.Join(dir, ".env")
if _, err := os.Stat(envPath); err == nil {
return envPath
}
parent := filepath.Dir(dir)
if parent == dir {
break
}
dir = parent
}

return ""
}

func validateConfig(cfg *Config) error {
if cfg.Redis.Credentials.Host == "" {
return fmt.Errorf("redis host cannot be empty")
}
if cfg.Redis.Credentials.Port <= 0 || cfg.Redis.Credentials.Port > 65535 {
return fmt.Errorf("redis port must be between 1 and 65535")
}
if cfg.Redis.Pool.MaxIdleConns < 0 {
return fmt.Errorf("max idle conns must be non-negative")
}
if cfg.Redis.Pool.MaxActiveConns < 0 {
return fmt.Errorf("max active conns must be non-negative")
}
if cfg.Redis.Connect.BackoffRetries < 0 {
return fmt.Errorf("backoff retries must be non-negative")
}
return nil
}

func (c *Config) GetRedisRetryConfig() *retry.Config {
strategy := retry.NewExponentialBackoff(retry.ExponentialBackoffConfig{
Min:    c.Redis.Connect.BackoffMin,
Max:    c.Redis.Connect.BackoffMax,
Factor: float64(c.Redis.Connect.BackoffFactor),
Jitter: c.Redis.Connect.BackoffJitter,
})

return &retry.Config{
MaxAttempts: c.Redis.Connect.BackoffRetries,
Strategy:    strategy,
}
}

func (c *Config) GetHost() string {
return c.Redis.Credentials.Host
}

func (c *Config) GetPort() int {
return c.Redis.Credentials.Port
}

func (c *Config) GetPassword() string {
return c.Redis.Credentials.Password
}

func (c *Config) GetDB() int {
return c.Redis.Credentials.DB
}

func (c *Config) GetMaxIdleConns() int {
return c.Redis.Pool.MaxIdleConns
}

func (c *Config) GetMaxActiveConns() int {
return c.Redis.Pool.MaxActiveConns
}

func (c *Config) GetQueryTimeout() time.Duration {
return c.Redis.Connect.QueryTimeout
}

func (c *Config) GetExecTimeout() time.Duration {
return c.Redis.Connect.ExecTimeout
}

func (c *Config) GetBackoffMin() time.Duration {
return c.Redis.Connect.BackoffMin
}

func (c *Config) GetBackoffMax() time.Duration {
return c.Redis.Connect.BackoffMax
}

func (c *Config) GetBackoffFactor() int {
return c.Redis.Connect.BackoffFactor
}

func (c *Config) GetBackoffJitter() bool {
return c.Redis.Connect.BackoffJitter
}

func (c *Config) GetBackoffRetries() int {
return c.Redis.Connect.BackoffRetries
}
