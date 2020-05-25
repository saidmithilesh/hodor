package gateway

import (
	"net/http"

	"github.com/saidmithilesh/hodor/config"
	"github.com/saidmithilesh/hodor/logging"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// Gateway type encapsulates all data, config and methods required
// for the API gteway to function. If using Hodor as a library in
// your own go project, just import the gateway package and create
// an instance of type Gateway. Then call the Build and Start methods
// on that instance to start using the gateway.
type Gateway struct {
	Router    *httprouter.Router
	Config    *config.Config
	Endpoints []*Endpoint
}

// Build method associates the gateway's config, sets up the router,
// and configures individual end points.
func (g *Gateway) Build(conf *config.Config) *Gateway {
	g.Config = conf
	g.Router = httprouter.New()

	for _, epc := range g.Config.Gateway.Endpoints {
		endpoint := NewEndpoint(&epc)
		endpoint.Build(g.Router)
		g.Endpoints = append(g.Endpoints, &endpoint)
	}

	return g
}

// Start method starts the http server using the router setup from
// the Build method.
func (g *Gateway) Start() {
	err := http.ListenAndServe(g.Config.Gateway.Port, g.Router)
	if err != nil {
		logging.Logger.Fatal(
			"Error while starting http server",
			zap.Error(err),
		)
	}
}
