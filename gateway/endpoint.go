package gateway

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/saidmithilesh/hodor/config"
	"github.com/saidmithilesh/hodor/logging"
	"go.uber.org/zap"

	uuid "github.com/satori/go.uuid"
)

// Endpoint data type
// A slice of instances of this type comprise the entire gateway
type Endpoint struct {
	Backend *url.URL
	Config  *config.EndpointConfig
}

func (e *Endpoint) proxyFunc(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	requestID := uuid.NewV4().String()

	logging.Logger.Info(
		"New request",
		zap.Uint("epid", e.Config.ID),
		zap.String("epname", e.Config.Name),
		zap.String("epmethod", e.Config.Method),
		zap.String("reqid", requestID),
	)

	req.Host = e.Backend.Host
	req.URL.Host = e.Backend.Host
	req.URL.Scheme = e.Backend.Scheme
	req.RequestURI = ""

	remoteAddr, _, _ := net.SplitHostPort(req.RemoteAddr)
	req.Header.Set("X-Forwarded-For", remoteAddr)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		logging.Logger.Info(
			"Request forwarding failed",
			zap.Uint("epid", e.Config.ID),
			zap.String("epname", e.Config.Name),
			zap.String("epmethod", e.Config.Method),
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

// Build the functionality for the endpoint
func (e *Endpoint) Build(r *httprouter.Router) *Endpoint {
	switch e.Config.Method {
	case http.MethodGet:
		r.GET(e.Config.Path, e.proxyFunc)
		return e

	case http.MethodPost:
		r.POST(e.Config.Path, e.proxyFunc)
		return e

	case http.MethodPut:
		r.PUT(e.Config.Path, e.proxyFunc)
		return e

	case http.MethodDelete:
		r.DELETE(e.Config.Path, e.proxyFunc)
		return e

	case http.MethodPatch:
		r.PATCH(e.Config.Path, e.proxyFunc)
		return e

	case http.MethodHead:
		r.HEAD(e.Config.Path, e.proxyFunc)
		return e

	case http.MethodOptions:
		r.OPTIONS(e.Config.Path, e.proxyFunc)
		return e

	default:
		logging.Logger.Fatal(
			"Error while configuring proxy function for endpoint",
			zap.Uint("epid", e.Config.ID),
			zap.String("epname", e.Config.Name),
			zap.String("epmethod", e.Config.Method),
		)
		return e
	}
}

// NewEndpoint creates an instance of type Endpoint and initiates
// it with the necessary configuration
func NewEndpoint(conf *config.EndpointConfig) Endpoint {
	var e Endpoint
	e.Config = conf
	remoteURL, err := url.Parse(e.Config.Backend)
	if err != nil {
		logging.Logger.Fatal(
			"Error while trying to parse backend url for endpoint",
			zap.Uint("epid", e.Config.ID),
			zap.String("epname", e.Config.Name),
			zap.String("epmethod", e.Config.Method),
		)
	}

	e.Backend = remoteURL
	return e
}
