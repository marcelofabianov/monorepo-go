package web

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	HTTP HTTPConfig
}

type HTTPConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	TLS          TLSConfig
	CORS         CORSConfig
	RateLimit    RateLimitConfig
}

type TLSConfig struct {
	Enabled  bool
	CertFile string
	KeyFile  string
}

type CORSConfig struct {
	Enabled          bool
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type RateLimitConfig struct {
	Enabled      bool
	RequestsPerSecond int
	Burst        int
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("WEB")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if envFile := findEnvFile(); envFile != "" {
		v.SetConfigFile(envFile)
		_ = v.ReadInConfig()
	}

	setDefaults(v)

	cfg := &Config{
		HTTP: HTTPConfig{
			Host:         v.GetString("http.host"),
			Port:         v.GetInt("http.port"),
			ReadTimeout:  v.GetDuration("http.read_timeout"),
			WriteTimeout: v.GetDuration("http.write_timeout"),
			IdleTimeout:  v.GetDuration("http.idle_timeout"),
			TLS: TLSConfig{
				Enabled:  v.GetBool("http.tls.enabled"),
				CertFile: v.GetString("http.tls.cert_file"),
				KeyFile:  v.GetString("http.tls.key_file"),
			},
			CORS: CORSConfig{
				Enabled:          v.GetBool("http.cors.enabled"),
				AllowedOrigins:   v.GetStringSlice("http.cors.allowed_origins"),
				AllowedMethods:   v.GetStringSlice("http.cors.allowed_methods"),
				AllowedHeaders:   v.GetStringSlice("http.cors.allowed_headers"),
				ExposedHeaders:   v.GetStringSlice("http.cors.exposed_headers"),
				AllowCredentials: v.GetBool("http.cors.allow_credentials"),
				MaxAge:           v.GetInt("http.cors.max_age"),
			},
			RateLimit: RateLimitConfig{
				Enabled:           v.GetBool("http.rate_limit.enabled"),
				RequestsPerSecond: v.GetInt("http.rate_limit.requests_per_second"),
				Burst:             v.GetInt("http.rate_limit.burst"),
			},
		},
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", 8080)
	v.SetDefault("http.read_timeout", 15*time.Second)
	v.SetDefault("http.write_timeout", 15*time.Second)
	v.SetDefault("http.idle_timeout", 60*time.Second)
	
	v.SetDefault("http.tls.enabled", false)
	v.SetDefault("http.tls.cert_file", "")
	v.SetDefault("http.tls.key_file", "")
	
	v.SetDefault("http.cors.enabled", true)
	v.SetDefault("http.cors.allowed_origins", []string{"*"})
	v.SetDefault("http.cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	v.SetDefault("http.cors.allowed_headers", []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"})
	v.SetDefault("http.cors.exposed_headers", []string{"X-Request-ID"})
	v.SetDefault("http.cors.allow_credentials", true)
	v.SetDefault("http.cors.max_age", 300)
	
	v.SetDefault("http.rate_limit.enabled", false)
	v.SetDefault("http.rate_limit.requests_per_second", 100)
	v.SetDefault("http.rate_limit.burst", 50)
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
