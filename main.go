package main

import (
	"github.com/saidmithilesh/hodor/config"
	"github.com/saidmithilesh/hodor/gateway"
	"github.com/saidmithilesh/hodor/logging"
)

func main() {
	// Load configuration from configuration file
	conf := config.LoadConfig()

	// Build the logger so that it can be imported and used
	// by other modules.
	logging.BuildLogger(&conf)

	// Construct the gateway object, build it and start
	// running it
	g := &gateway.Gateway{}
	g = g.Build(&conf)
	g.Start()
}
