package logging

import (
	"log"

	"go.uber.org/zap"

	"github.com/saidmithilesh/hodor/config"
)

// Logger is a singleton instance of type zap.Logger
var Logger *zap.Logger

// BuildLogger constructs the logger instance and initiates it with
// gateway config to make it possible to identify the gateway instance
// uniquely across all deployed instances. Since the gateway is designed
// to be deployed in an autoscaling fashion and all instances will send
// their logs to a centralised location, it is important to be able
// to distinguish between the logs produced by each instance to be able
// pinpoint the source of each log line.
func BuildLogger(conf *config.Config) {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.InitialFields = map[string]interface{}{
		"gatewayId":   conf.Gateway.ID,
		"instanceId":  conf.Gateway.InstanceID,
		"gatewayName": conf.Gateway.Name,
	}

	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatalf("Error while instantiating logger :: %#v", err)
	}

	Logger = logger
	Logger.Info("Logger initialised")
}
