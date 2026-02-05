package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds the logger configuration
type Config struct {
	Level       LogLevel
	Format      LogFormat
	Output      io.Writer
	ServiceName string
	Environment string
	AddSource   bool
	TimeFormat  string
}

// LoadConfig loads logger configuration from environment variables using Viper.
// It looks for a .env file in the current directory and up to 5 parent directories.
func LoadConfig() (*Config, error) {
	v := viper.New()

	// Find and load .env file
	envFile := findEnvFile()
	if envFile != "" {
		v.SetConfigFile(envFile)
		v.SetConfigType("env")
		_ = v.ReadInConfig() // Ignore error, we have defaults
	}

	// Environment variables take precedence
	v.AutomaticEnv()
	v.SetEnvPrefix("LOGGER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults(v)

	// Build config
	cfg := &Config{
		Level:       parseLevel(v.GetString("level")),
		Format:      determineFormat(v.GetString("environment")),
		Output:      os.Stdout,
		ServiceName: v.GetString("service_name"),
		Environment: v.GetString("environment"),
		AddSource:   shouldAddSource(v.GetString("environment")),
		TimeFormat:  time.RFC3339,
	}

	return cfg, nil
}

// setDefaults configures default values
func setDefaults(v *viper.Viper) {
	v.SetDefault("level", "info")
	v.SetDefault("environment", "development")
	v.SetDefault("service_name", "app")
}

// findEnvFile searches for .env file in current and parent directories (up to 5 levels)
func findEnvFile() string {
	dir, _ := os.Getwd()
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
	return ".env" // fallback
}

// parseLevel converts string log level to LogLevel
func parseLevel(level string) LogLevel {
	// Remove quotes if present
	level = strings.Trim(level, `"`)

	switch strings.ToLower(level) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// determineFormat returns the appropriate log format based on environment
func determineFormat(env string) LogFormat {
	env = strings.ToLower(env)
	if env == "production" || env == "prod" || env == "staging" {
		return FormatJSON
	}
	return FormatText
}

// shouldAddSource determines if source location should be added to logs
func shouldAddSource(env string) bool {
	env = strings.ToLower(env)
	return env == "development" || env == "dev"
}
