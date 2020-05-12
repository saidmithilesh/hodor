package server

import (
	"net/http"
	"sync"

	"github.com/saidmithilesh/hodor/utils"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

var once sync.Once

// HTTPServer struct provides methods for building and running an HTTP server
// Creating multiple instance of this struct will allow the application to run
// multiple HTTP servers, each listening on a port specified by the config
type HTTPServer struct {
	Router *httprouter.Router
}

// Build method instantiates the HTTP server and configures the routes based
// on the configuration provided. Each route acts as a proxy end point and
// implements its own logic to handle the HTTP requests going through.
func (srv *HTTPServer) Build(config *utils.Config) {
	once.Do(func() {
		srv.Router = httprouter.New()

		BuildEndpoints(srv.Router, config)
	})
}

// Start method causes the HTTP server to begin listening on the port provided
// by the configuration
func (srv *HTTPServer) Start(config *utils.Config) {
	err := http.ListenAndServe(config.Gateway.Port, srv.Router)
	if err != nil {
		utils.Logger.Fatal(
			"Error while starting http server",
			zap.Error(err),
		)
	}
}
