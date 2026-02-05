package logger_test

import (
	"testing"

	"github.com/marcelofabianov/logger"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Test with default values (no .env file in test context)
	cfg, err := logger.LoadConfig()
	
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, logger.LevelInfo, cfg.Level)
	assert.Equal(t, "development", cfg.Environment)
	assert.Equal(t, "app", cfg.ServiceName)
	assert.Equal(t, logger.FormatText, cfg.Format)
	assert.True(t, cfg.AddSource)
}

func TestLoadConfigWithLogger(t *testing.T) {
	// Load config and create logger
	cfg, err := logger.LoadConfig()
	assert.NoError(t, err)
	
	log := logger.New(cfg)
	assert.NotNil(t, log)
	
	// Test logger works
	log.Info("test message")
	log.Debug("debug message")
}
