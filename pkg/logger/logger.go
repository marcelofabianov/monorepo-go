package logger

import (
	"context"
	"log/slog"
	"os"
	"time"
)

type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

type LogFormat string

const (
	FormatJSON LogFormat = "json"
	FormatText LogFormat = "text"
)

type Logger struct {
	logger      *slog.Logger
	config      *Config
	serviceName string
	environment string
}

func New(cfg *Config) *Logger {
	if cfg == nil {
		cfg = defaultConfig()
	}

	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "unknown-service"
	}
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}

	level := parseLogLevel(cfg.Level)

	handlerOpts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(cfg.TimeFormat))
				}
			}
			return a
		},
	}

	var handler slog.Handler
	switch cfg.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(cfg.Output, handlerOpts)
	case FormatText:
		handler = slog.NewTextHandler(cfg.Output, handlerOpts)
	default:
		handler = slog.NewJSONHandler(cfg.Output, handlerOpts)
	}

	baseLogger := slog.New(handler)
	baseLogger = baseLogger.With(
		slog.String("service", cfg.ServiceName),
		slog.String("environment", cfg.Environment),
	)

	return &Logger{
		logger:      baseLogger,
		config:      cfg,
		serviceName: cfg.ServiceName,
		environment: cfg.Environment,
	}
}

func defaultConfig() *Config {
	return &Config{
		Level:       LevelInfo,
		Format:      FormatJSON,
		Output:      os.Stdout,
		ServiceName: "unknown-service",
		Environment: "development",
		AddSource:   false,
		TimeFormat:  time.RFC3339,
	}
}

func parseLogLevel(level LogLevel) slog.Level {
	switch level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.logger.DebugContext(ctx, msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.logger.InfoContext(ctx, msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.logger.WarnContext(ctx, msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.logger.ErrorContext(ctx, msg, args...)
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		logger:      l.logger.With(args...),
		config:      l.config,
		serviceName: l.serviceName,
		environment: l.environment,
	}
}

func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		logger:      l.logger.WithGroup(name),
		config:      l.config,
		serviceName: l.serviceName,
		environment: l.environment,
	}
}

func (l *Logger) Slog() *slog.Logger {
	return l.logger
}

func (l *Logger) SetDefault() {
	slog.SetDefault(l.logger)
}

func (l *Logger) GetConfig() Config {
	return *l.config
}

func (l *Logger) ServiceName() string {
	return l.serviceName
}

func (l *Logger) Environment() string {
	return l.environment
}

func (l *Logger) Enabled(ctx context.Context, level LogLevel) bool {
	slogLevel := parseLogLevel(level)
	return l.logger.Enabled(ctx, slogLevel)
}

func (l *Logger) LogAttrs(ctx context.Context, level LogLevel, msg string, attrs ...slog.Attr) {
	slogLevel := parseLogLevel(level)
	l.logger.LogAttrs(ctx, slogLevel, msg, attrs...)
}

func (l *Logger) Handler() slog.Handler {
	return l.logger.Handler()
}

func NewFromSlog(slogger *slog.Logger, serviceName, environment string) *Logger {
	return &Logger{
		logger:      slogger,
		config:      nil,
		serviceName: serviceName,
		environment: environment,
	}
}
