package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("cria logger com configurações padrão", func(t *testing.T) {
		logger := New(nil)

		assert.NotNil(t, logger)
		assert.NotNil(t, logger.logger)
		assert.Equal(t, "unknown-service", logger.serviceName)
		assert.Equal(t, "development", logger.environment)
	})

	t.Run("cria logger com configurações customizadas", func(t *testing.T) {
		cfg := &Config{
			Level:       LevelDebug,
			Format:      FormatJSON,
			ServiceName: "test-service",
			Environment: "test",
			AddSource:   true,
		}

		logger := New(cfg)

		assert.NotNil(t, logger)
		assert.Equal(t, "test-service", logger.serviceName)
		assert.Equal(t, "test", logger.environment)
	})

	t.Run("aplica valores padrão quando não fornecidos", func(t *testing.T) {
		cfg := &Config{}
		logger := New(cfg)

		assert.NotNil(t, logger.config.Output)
		assert.Equal(t, time.RFC3339, logger.config.TimeFormat)
		assert.Equal(t, "unknown-service", logger.serviceName)
		assert.Equal(t, "development", logger.environment)
	})
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name     string
		logLevel LogLevel
		logFunc  func(*Logger, string)
		expected string
	}{
		{
			name:     "debug level",
			logLevel: LevelDebug,
			logFunc:  func(l *Logger, msg string) { l.Debug(msg) },
			expected: "DEBUG",
		},
		{
			name:     "info level",
			logLevel: LevelInfo,
			logFunc:  func(l *Logger, msg string) { l.Info(msg) },
			expected: "INFO",
		},
		{
			name:     "warn level",
			logLevel: LevelWarn,
			logFunc:  func(l *Logger, msg string) { l.Warn(msg) },
			expected: "WARN",
		},
		{
			name:     "error level",
			logLevel: LevelError,
			logFunc:  func(l *Logger, msg string) { l.Error(msg) },
			expected: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cfg := &Config{
				Level:       LevelDebug,
				Format:      FormatText,
				Output:      &buf,
				ServiceName: "test",
				Environment: "test",
			}

			logger := New(cfg)
			tt.logFunc(logger, "test message")

			output := buf.String()
			assert.Contains(t, output, tt.expected)
			assert.Contains(t, output, "test message")
		})
	}
}

func TestLogFormat(t *testing.T) {
	t.Run("formato JSON", func(t *testing.T) {
		var buf bytes.Buffer
		cfg := &Config{
			Level:       LevelInfo,
			Format:      FormatJSON,
			Output:      &buf,
			ServiceName: "test-service",
			Environment: "test",
		}

		logger := New(cfg)
		logger.Info("test message", "key", "value")

		output := buf.String()

		var jsonLog map[string]interface{}
		err := json.Unmarshal([]byte(output), &jsonLog)
		require.NoError(t, err)

		assert.Equal(t, "test message", jsonLog["msg"])
		assert.Equal(t, "INFO", jsonLog["level"])
		assert.Equal(t, "test-service", jsonLog["service"])
		assert.Equal(t, "test", jsonLog["environment"])
		assert.Equal(t, "value", jsonLog["key"])
		assert.NotEmpty(t, jsonLog["time"])
	})

	t.Run("formato Text", func(t *testing.T) {
		var buf bytes.Buffer
		cfg := &Config{
			Level:       LevelInfo,
			Format:      FormatText,
			Output:      &buf,
			ServiceName: "test-service",
			Environment: "test",
		}

		logger := New(cfg)
		logger.Info("test message", "key", "value")

		output := buf.String()

		assert.Contains(t, output, "test message")
		assert.Contains(t, output, "INFO")
		assert.Contains(t, output, "service=test-service")
		assert.Contains(t, output, "environment=test")
		assert.Contains(t, output, "key=value")
	})
}

func TestLogWithFields(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		Output:      &buf,
		ServiceName: "test",
		Environment: "test",
	}

	logger := New(cfg)
	logger.Info("test message",
		"user_id", "123",
		"action", "login",
		"duration_ms", 150,
	)

	var jsonLog map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &jsonLog)
	require.NoError(t, err)

	assert.Equal(t, "test message", jsonLog["msg"])
	assert.Equal(t, "123", jsonLog["user_id"])
	assert.Equal(t, "login", jsonLog["action"])
	assert.Equal(t, float64(150), jsonLog["duration_ms"])
}

func TestLogWithContext(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Level:       LevelInfo,
		Format:      FormatText,
		Output:      &buf,
		ServiceName: "test",
		Environment: "test",
	}

	logger := New(cfg)
	ctx := context.Background()

	logger.InfoContext(ctx, "test message with context")

	output := buf.String()
	assert.Contains(t, output, "test message with context")
	assert.Contains(t, output, "INFO")
}

func TestWith(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		Output:      &buf,
		ServiceName: "test",
		Environment: "test",
	}

	logger := New(cfg)

	requestLogger := logger.With(
		"request_id", "req-123",
		"user_id", "user-456",
	)

	requestLogger.Info("processing request")

	var jsonLog map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &jsonLog)
	require.NoError(t, err)

	assert.Equal(t, "processing request", jsonLog["msg"])
	assert.Equal(t, "req-123", jsonLog["request_id"])
	assert.Equal(t, "user-456", jsonLog["user_id"])
	assert.Equal(t, "test", jsonLog["service"])
	assert.Equal(t, "test", jsonLog["environment"])
}

func TestWithGroup(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		Output:      &buf,
		ServiceName: "test",
		Environment: "test",
	}

	logger := New(cfg)
	groupLogger := logger.WithGroup("request")

	groupLogger.Info("handling request",
		"method", "GET",
		"path", "/api/users",
	)

	var jsonLog map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &jsonLog)
	require.NoError(t, err)

	requestGroup, ok := jsonLog["request"].(map[string]interface{})
	require.True(t, ok, "request group should exist")
	assert.Equal(t, "GET", requestGroup["method"])
	assert.Equal(t, "/api/users", requestGroup["path"])
}

func TestLogLevelFiltering(t *testing.T) {
	tests := []struct {
		name           string
		configLevel    LogLevel
		logLevel       LogLevel
		shouldBeLogged bool
	}{
		{"debug level logs debug", LevelDebug, LevelDebug, true},
		{"info level filters debug", LevelInfo, LevelDebug, false},
		{"info level logs info", LevelInfo, LevelInfo, true},
		{"warn level filters info", LevelWarn, LevelInfo, false},
		{"warn level logs warn", LevelWarn, LevelWarn, true},
		{"error level filters warn", LevelError, LevelWarn, false},
		{"error level logs error", LevelError, LevelError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cfg := &Config{
				Level:       tt.configLevel,
				Format:      FormatText,
				Output:      &buf,
				ServiceName: "test",
				Environment: "test",
			}

			logger := New(cfg)

			switch tt.logLevel {
			case LevelDebug:
				logger.Debug("debug message")
			case LevelInfo:
				logger.Info("info message")
			case LevelWarn:
				logger.Warn("warn message")
			case LevelError:
				logger.Error("error message")
			}

			output := buf.String()
			if tt.shouldBeLogged {
				assert.NotEmpty(t, output, "log should have been written")
			} else {
				assert.Empty(t, output, "log should have been filtered")
			}
		})
	}
}

func TestEnabled(t *testing.T) {
	cfg := &Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		Output:      &bytes.Buffer{},
		ServiceName: "test",
		Environment: "test",
	}

	logger := New(cfg)
	ctx := context.Background()

	assert.False(t, logger.Enabled(ctx, LevelDebug))
	assert.True(t, logger.Enabled(ctx, LevelInfo))
	assert.True(t, logger.Enabled(ctx, LevelWarn))
	assert.True(t, logger.Enabled(ctx, LevelError))
}

func TestSetDefault(t *testing.T) {
	cfg := &Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		Output:      &bytes.Buffer{},
		ServiceName: "test",
		Environment: "test",
	}

	logger := New(cfg)
	logger.SetDefault()

	defaultLogger := slog.Default()
	assert.NotNil(t, defaultLogger)
}

func TestSlog(t *testing.T) {
	cfg := &Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		Output:      &bytes.Buffer{},
		ServiceName: "test",
		Environment: "test",
	}

	logger := New(cfg)
	slogger := logger.Slog()

	assert.NotNil(t, slogger)
	assert.IsType(t, &slog.Logger{}, slogger)
}

func TestGetConfig(t *testing.T) {
	originalCfg := &Config{
		Level:       LevelDebug,
		Format:      FormatJSON,
		ServiceName: "test-service",
		Environment: "production",
		AddSource:   true,
	}

	logger := New(originalCfg)
	cfg := logger.GetConfig()

	assert.Equal(t, LevelDebug, cfg.Level)
	assert.Equal(t, FormatJSON, cfg.Format)
	assert.Equal(t, "test-service", cfg.ServiceName)
	assert.Equal(t, "production", cfg.Environment)
	assert.True(t, cfg.AddSource)
}

func TestServiceNameAndEnvironment(t *testing.T) {
	cfg := &Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		Output:      &bytes.Buffer{},
		ServiceName: "my-service",
		Environment: "staging",
	}

	logger := New(cfg)

	assert.Equal(t, "my-service", logger.ServiceName())
	assert.Equal(t, "staging", logger.Environment())
}

func TestNewFromSlog(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	slogger := slog.New(handler)

	logger := NewFromSlog(slogger, "test-service", "test")

	assert.NotNil(t, logger)
	assert.Equal(t, "test-service", logger.ServiceName())
	assert.Equal(t, "test", logger.Environment())
	assert.Nil(t, logger.config)

	logger.Info("test message")
	assert.Contains(t, buf.String(), "test message")
}

func TestLogAttrs(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		Output:      &buf,
		ServiceName: "test",
		Environment: "test",
	}

	logger := New(cfg)
	ctx := context.Background()

	attrs := []slog.Attr{
		slog.String("user_id", "123"),
		slog.Int("count", 42),
		slog.Bool("active", true),
	}

	logger.LogAttrs(ctx, LevelInfo, "test with attrs", attrs...)

	var jsonLog map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &jsonLog)
	require.NoError(t, err)

	assert.Equal(t, "test with attrs", jsonLog["msg"])
	assert.Equal(t, "123", jsonLog["user_id"])
	assert.Equal(t, float64(42), jsonLog["count"])
	assert.Equal(t, true, jsonLog["active"])
}

func TestHandler(t *testing.T) {
	cfg := &Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		Output:      &bytes.Buffer{},
		ServiceName: "test",
		Environment: "test",
	}

	logger := New(cfg)
	handler := logger.Handler()

	assert.NotNil(t, handler)
	assert.Implements(t, (*slog.Handler)(nil), handler)
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    LogLevel
		expected slog.Level
	}{
		{LevelDebug, slog.LevelDebug},
		{LevelInfo, slog.LevelInfo},
		{LevelWarn, slog.LevelWarn},
		{LevelError, slog.LevelError},
		{LogLevel("invalid"), slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result := parseLogLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAddSource(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		Output:      &buf,
		ServiceName: "test",
		Environment: "test",
		AddSource:   true,
	}

	logger := New(cfg)
	logger.Info("test with source")

	output := buf.String()
	assert.Contains(t, output, "source")
}

func TestCustomTimeFormat(t *testing.T) {
	var buf bytes.Buffer
	customFormat := "2006-01-02 15:04:05"
	cfg := &Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		Output:      &buf,
		ServiceName: "test",
		Environment: "test",
		TimeFormat:  customFormat,
	}

	logger := New(cfg)
	logger.Info("test message")

	var jsonLog map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &jsonLog)
	require.NoError(t, err)

	timeStr, ok := jsonLog["time"].(string)
	require.True(t, ok)

	_, err = time.Parse(customFormat, timeStr)
	assert.NoError(t, err, "time should be in custom format")
}

func BenchmarkLogger(b *testing.B) {
	cfg := &Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		Output:      &bytes.Buffer{},
		ServiceName: "benchmark",
		Environment: "test",
	}

	logger := New(cfg)

	b.Run("Info without fields", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message")
		}
	})

	b.Run("Info with fields", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message",
				"user_id", "123",
				"action", "test",
				"count", i,
			)
		}
	})

	b.Run("With + Info", func(b *testing.B) {
		childLogger := logger.With("request_id", "req-123")
		for i := 0; i < b.N; i++ {
			childLogger.Info("benchmark message")
		}
	})
}

func TestConcurrency(t *testing.T) {
	var buf bytes.Buffer
	cfg := &Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		Output:      &buf,
		ServiceName: "test",
		Environment: "test",
	}

	logger := New(cfg)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Info("concurrent log", "goroutine", id)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	assert.Equal(t, 10, len(lines))
}
