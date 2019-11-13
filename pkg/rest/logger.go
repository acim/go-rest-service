package rest

import (
	"fmt"

	"go.uber.org/zap"
)

// NewLogger creates new zap logger.
func NewLogger(env string) (*zap.Logger, error) {
	var logger *zap.Logger

	var err error

	switch env {
	case "prod":
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
		logger, err = config.Build()
	case "dev":
		logger, err = zap.NewDevelopment()
	default:
		return nil, fmt.Errorf("logger: unknown environment: '%s'", env)
	}

	if err != nil {
		return nil, fmt.Errorf("logger: %w", err)
	}

	return logger, nil
}
