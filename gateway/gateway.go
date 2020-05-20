package gateway

import (
	"net/http"

	"github.com/saidmithilesh/hodor/config"
	"github.com/saidmithilesh/hodor/logging"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

type Gateway struct {
	Router    *httprouter.Router
	Config    *config.Config
	Endpoints []*Endpoint
}

func (g *Gateway) Build(conf *config.Config) *Gateway {
	g.Router = httprouter.New()
	g.Config = conf

	for _, epc := range g.Config.Gateway.Endpoints {
		endpoint := NewEndpoint(&epc)
		endpoint.Build(g.Router)
		g.Endpoints = append(g.Endpoints, &endpoint)
	}

	return g
}

func (g *Gateway) Start() {
	err := http.ListenAndServe(g.Config.Gateway.Port, g.Router)
	if err != nil {
		logging.Logger.Fatal(
			"Error while starting http server",
			zap.Error(err),
		)
	}
}
