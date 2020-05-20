package main

import (
	"github.com/saidmithilesh/hodor/config"
	"github.com/saidmithilesh/hodor/gateway"
	"github.com/saidmithilesh/hodor/logging"
)

func main() {
	conf := config.LoadConfig()
	logging.BuildLogger(&conf)
	g := &gateway.Gateway{}
	g = g.Build(&conf)
	g.Start()
}
