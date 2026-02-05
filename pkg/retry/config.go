package retry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type BackoffConfig struct {
	Type      string
	Min       time.Duration
	Max       time.Duration
	Factor    float64
	Jitter    bool
	Delay     time.Duration
	Increment time.Duration
}

type RetryConfig struct {
	MaxAttempts int
	Backoff     BackoffConfig
}

func LoadConfig() *RetryConfig {
	v := viper.New()
	v.SetEnvPrefix("RETRY")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if envFile := findEnvFile(); envFile != "" {
		v.SetConfigFile(envFile)
		_ = v.ReadInConfig()
	}

	setDefaults(v)

	return &RetryConfig{
		MaxAttempts: v.GetInt("max_attempts"),
		Backoff: BackoffConfig{
			Type:      v.GetString("backoff.type"),
			Min:       v.GetDuration("backoff.min"),
			Max:       v.GetDuration("backoff.max"),
			Factor:    v.GetFloat64("backoff.factor"),
			Jitter:    v.GetBool("backoff.jitter"),
			Delay:     v.GetDuration("backoff.delay"),
			Increment: v.GetDuration("backoff.increment"),
		},
	}
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("max_attempts", 3)
	v.SetDefault("backoff.type", "exponential")
	v.SetDefault("backoff.min", 1*time.Second)
	v.SetDefault("backoff.max", 30*time.Second)
	v.SetDefault("backoff.factor", 2.0)
	v.SetDefault("backoff.jitter", true)
	v.SetDefault("backoff.delay", 1*time.Second)
	v.SetDefault("backoff.increment", 1*time.Second)
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

func (bc *BackoffConfig) CreateStrategy() (Strategy, error) {
	switch bc.Type {
	case "exponential":
		return NewExponentialBackoff(ExponentialBackoffConfig{
			Min:    bc.Min,
			Max:    bc.Max,
			Factor: bc.Factor,
			Jitter: bc.Jitter,
		}), nil

	case "constant":
		return NewConstantBackoff(bc.Delay), nil

	case "linear":
		return NewLinearBackoff(bc.Increment, bc.Max), nil

	default:
		return nil, fmt.Errorf("unknown backoff type: %s", bc.Type)
	}
}

func (rc *RetryConfig) ToConfig() (*Config, error) {
	strategy, err := rc.Backoff.CreateStrategy()
	if err != nil {
		return nil, err
	}

	return &Config{
		MaxAttempts: rc.MaxAttempts,
		Strategy:    strategy,
	}, nil
}
