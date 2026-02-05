package logger

import (
	"fmt"
	"os"
)

func ExampleNew() {
	cfg := &Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		ServiceName: "my-api",
		Environment: "production",
	}

	logger := New(cfg)
	logger.Info("application started", "port", 8080)
}

func ExampleLogger_With() {
	logger := New(&Config{
		Level:       LevelInfo,
		Format:      FormatText,
		Output:      os.Stdout,
		ServiceName: "api",
		Environment: "dev",
	})

	requestLogger := logger.With(
		"request_id", "req-123",
		"user_id", "user-456",
	)

	requestLogger.Info("processing request")
	requestLogger.Info("request completed", "duration_ms", 150)
}

func ExampleLogger_WithGroup() {
	logger := New(&Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		ServiceName: "api",
		Environment: "dev",
	})

	httpLogger := logger.WithGroup("http")
	httpLogger.Info("request received",
		"method", "GET",
		"path", "/api/users",
		"status", 200,
	)
}

func ExampleLogger_Debug() {
	logger := New(&Config{
		Level:       LevelDebug,
		Format:      FormatText,
		ServiceName: "api",
		Environment: "dev",
	})

	logger.Debug("detailed debugging information",
		"query", "SELECT * FROM users",
		"params", []int{1, 2, 3},
	)
}

func ExampleLogger_Info() {
	logger := New(&Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		ServiceName: "api",
		Environment: "prod",
	})

	logger.Info("user logged in",
		"user_id", "123",
		"ip", "192.168.1.1",
	)
}

func ExampleLogger_Warn() {
	logger := New(&Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		ServiceName: "api",
		Environment: "prod",
	})

	logger.Warn("rate limit approaching threshold",
		"current", 95,
		"limit", 100,
		"user_id", "123",
	)
}

func ExampleLogger_Error() {
	logger := New(&Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		ServiceName: "api",
		Environment: "prod",
	})

	err := fmt.Errorf("database connection failed")
	logger.Error("failed to connect to database",
		"error", err,
		"host", "localhost",
		"port", 5432,
	)
}

func ExampleLogger_SetDefault() {
	logger := New(&Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		ServiceName: "api",
		Environment: "prod",
	})

	logger.SetDefault()
}
