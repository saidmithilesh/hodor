package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/saidmithilesh/hodor/utils"
	"go.uber.org/zap"
)

// Endpoint - instances of this structure hold the configuration and
// functionality required to manage an individual end point
type Endpoint struct {
	config     *utils.EndpointConfig
	backendURL *url.URL
}

// TODO: Add custom response fields to the endpoint config and accordingly write
//       the response depending on each case
func (e *Endpoint) proxyFunc(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	req.Host = e.backendURL.Host
	req.URL.Host = e.backendURL.Host
	req.URL.Scheme = e.backendURL.Scheme
	req.RequestURI = ""

	remoteAddr, _, _ := net.SplitHostPort(req.RemoteAddr)
	req.Header.Set("X-Forwarded-For", remoteAddr)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		utils.Logger.Info(
			"Request forwarding failed",
			zap.Uint("endpointId", e.config.ID),
			zap.String("endpointName", e.config.Name),
			zap.Error(err),
		)
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(res, "Internal server error")
		return
	}

	// Copy all the headers received from the backend
	for key, values := range response.Header {
		for _, value := range values {
			res.Header().Set(key, value)
		}
	}
	res.WriteHeader(response.StatusCode)
	io.Copy(res, response.Body)
}

func (e *Endpoint) build(r *httprouter.Router) error {
	switch e.config.Method {

	case http.MethodGet:
		r.GET(e.config.Path, e.proxyFunc)
		return nil

	case http.MethodPost:
		r.POST(e.config.Path, e.proxyFunc)
		return nil

	case http.MethodPut:
		r.PUT(e.config.Path, e.proxyFunc)
		return nil

	case http.MethodDelete:
		r.DELETE(e.config.Path, e.proxyFunc)
		return nil

	case http.MethodHead:
		r.HEAD(e.config.Path, e.proxyFunc)
		return nil

	case http.MethodOptions:
		r.OPTIONS(e.config.Path, e.proxyFunc)
		return nil

	case http.MethodPatch:
		r.PATCH(e.config.Path, e.proxyFunc)
		return nil

	default:
		return errors.New(
			fmt.Sprintf(
				"Invalid method '%s' specified for endpoint '%s'",
				e.config.Method,
				e.config.Name,
			),
		)
	}
}

// BuildEndpoints function is called from the httpserver module
// It iterates over all the endpoints specified in the configuration
// and builds a proxy function for them based on their type and settings
func BuildEndpoints(r *httprouter.Router, config *utils.Config) {
	for _, epc := range config.Gateway.Endpoints {
		endpoint := Endpoint{}
		endpoint.config = &epc
		remoteURL, err := url.Parse(endpoint.config.Backend)
		if err != nil {
			utils.Logger.Fatal(
				"Error while parsing remote URL for specified backend",
				zap.Uint("endpointId", epc.ID),
				zap.String("endpointName", epc.Name),
				zap.Error(err),
			)
			log.Fatalf("Error while parsing remote URL %#v", err)
		}
		endpoint.backendURL = remoteURL

		err = endpoint.build(r)
		if err != nil {
			log.Fatalf("Error while building endpoint %s :: %#v", endpoint.config.Name, err)
		}
	}
}
