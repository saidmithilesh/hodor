package utils

import (
	"log"

	"go.uber.org/zap"
)

// Logger is the global logger instance that will be shared and used by all
// modules for logging purposes
var Logger *zap.Logger

// BuildLogger builds a logger object and returns it back to the main function
func BuildLogger(config *Config) {
	cfg := zap.NewProductionConfig()
	cfg.InitialFields = map[string]interface{}{
		"gatewayId":   config.Gateway.ID,
		"instanceId":  config.Gateway.InstanceID,
		"gatewayName": config.Gateway.Name,
	}

	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("Error while instantiating logger :: %#v", err)
	}

	Logger = logger
	Logger.Info("Logger initialised")
}

// GetLogger returns an instance of the logger type
func GetLogger() *zap.Logger {
	return Logger
}
