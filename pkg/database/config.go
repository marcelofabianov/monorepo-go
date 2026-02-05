package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Credentials DatabaseCredentialsConfig
	Connect     DatabaseConnectConfig
	Pool        DatabasePoolConfig
}

type DatabaseCredentialsConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type DatabaseConnectConfig struct {
	QueryTimeout   time.Duration
	ExecTimeout    time.Duration
	BackoffMin     time.Duration
	BackoffMax     time.Duration
	BackoffFactor  int
	BackoffJitter  bool
	BackoffRetries int
}

type DatabasePoolConfig struct {
	MaxOpenConns      int
	MaxIdleConns      int
	ConnMaxLifetime   time.Duration
	ConnMaxIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("DATABASE")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if envFile := findEnvFile(); envFile != "" {
		v.SetConfigFile(envFile)
		_ = v.ReadInConfig()
	}

	setDefaults(v)

	cfg := &Config{
		Database: DatabaseConfig{
			Credentials: DatabaseCredentialsConfig{
				Host:     v.GetString("host"),
				Port:     v.GetInt("port"),
				User:     v.GetString("user"),
				Password: v.GetString("password"),
				Name:     v.GetString("name"),
				SSLMode:  v.GetString("sslmode"),
			},
			Connect: DatabaseConnectConfig{
				QueryTimeout:   v.GetDuration("connect.query_timeout"),
				ExecTimeout:    v.GetDuration("connect.exec_timeout"),
				BackoffMin:     v.GetDuration("connect.backoff_min"),
				BackoffMax:     v.GetDuration("connect.backoff_max"),
				BackoffFactor:  v.GetInt("connect.backoff_factor"),
				BackoffJitter:  v.GetBool("connect.backoff_jitter"),
				BackoffRetries: v.GetInt("connect.backoff_retries"),
			},
			Pool: DatabasePoolConfig{
				MaxOpenConns:      v.GetInt("pool.max_open_conns"),
				MaxIdleConns:      v.GetInt("pool.max_idle_conns"),
				ConnMaxLifetime:   v.GetDuration("pool.conn_max_lifetime"),
				ConnMaxIdleTime:   v.GetDuration("pool.conn_max_idle_time"),
				HealthCheckPeriod: v.GetDuration("pool.health_check_period"),
			},
		},
	}

	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("host", "localhost")
	v.SetDefault("port", 5432)
	v.SetDefault("user", "postgres")
	v.SetDefault("password", "")
	v.SetDefault("name", "postgres")
	v.SetDefault("sslmode", "disable")
	v.SetDefault("connect.query_timeout", 5*time.Second)
	v.SetDefault("connect.exec_timeout", 10*time.Second)
	v.SetDefault("connect.backoff_min", 500*time.Millisecond)
	v.SetDefault("connect.backoff_max", 30*time.Second)
	v.SetDefault("connect.backoff_factor", 2)
	v.SetDefault("connect.backoff_jitter", true)
	v.SetDefault("connect.backoff_retries", 5)
	v.SetDefault("pool.max_open_conns", 25)
	v.SetDefault("pool.max_idle_conns", 5)
	v.SetDefault("pool.conn_max_lifetime", 5*time.Minute)
	v.SetDefault("pool.conn_max_idle_time", 5*time.Minute)
	v.SetDefault("pool.health_check_period", 30*time.Second)
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

func ValidateConfig(cfg *Config) error {
	if cfg.Database.Credentials.Host == "" {
		return fmt.Errorf("database host cannot be empty")
	}
	if cfg.Database.Credentials.Port <= 0 || cfg.Database.Credentials.Port > 65535 {
		return fmt.Errorf("database port must be between 1 and 65535")
	}
	if cfg.Database.Credentials.User == "" {
		return fmt.Errorf("database user cannot be empty")
	}
	if cfg.Database.Credentials.Name == "" {
		return fmt.Errorf("database name cannot be empty")
	}
	if cfg.Database.Pool.MaxOpenConns < 1 {
		return fmt.Errorf("max open conns must be at least 1")
	}
	if cfg.Database.Pool.MaxIdleConns < 0 {
		return fmt.Errorf("max idle conns must be non-negative")
	}
	if cfg.Database.Connect.BackoffRetries < 0 {
		return fmt.Errorf("backoff retries must be non-negative")
	}
	return nil
}

func (c *Config) GetDatabaseDSN() string {
	creds := c.Database.Credentials
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		creds.Host,
		creds.Port,
		creds.User,
		creds.Password,
		creds.Name,
		creds.SSLMode,
	)
}
