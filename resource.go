package api

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/service-status-go/buildinfo"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"

	"gopkg.in/yaml.v2"
)

// DefaultPath is the expected path for the Endpoint to be served at
const DefaultPath = "/__api"

// Endpoint provides an API http endpoint which should be served on the DefaultPath
type Endpoint interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type endpoint struct {
	yml       []byte
	parsedAPI map[string]interface{}
	buildInfo buildinfo.BuildInfo
}

// NewAPIEndpointForFile reads the swagger yml file at the provided location, and returns an Endpoint
func NewAPIEndpointForFile(apiFile string) (Endpoint, error) {
	file, err := ioutil.ReadFile(apiFile)
	if err != nil {
		return nil, err
	}

	return NewAPIEndpointForYAML(file)
}

// NewAPIEndpointForYAML returns an endpoint for the given swagger yml as a []byte
func NewAPIEndpointForYAML(yml []byte) (Endpoint, error) {
	api := make(map[string]interface{})
	err := yaml.Unmarshal(yml, &api)
	if err != nil {
		return nil, err
	}

	build := buildinfo.GetBuildInfo()

	return &endpoint{yml: yml, parsedAPI: api, buildInfo: build}, nil
}

// GetEndpoint returns the endpoint which handles API request and amends the relevant fields dynamically
func (e *endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uri := r.Header.Get("X-Original-Request-URL")
	if strings.TrimSpace(uri) == "" {
		w.Write(e.yml)
		return
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		w.Write(e.yml)
		return
	}

	api := copyMap(e.parsedAPI)

	api["host"] = parsed.Host
	api["schemes"] = []string{"https"}
	api["basePath"] = basePath(parsed.Path)

	info, ok := api["info"].(map[interface{}]interface{})
	if ok {
		info["version"] = e.buildInfo.Version
	}

	out, err := yaml.Marshal(api)
	if err != nil {
		w.Write(e.yml)
		return
	}

	w.Write(out)
}

func copyMap(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

func basePath(path string) string {
	if strings.HasSuffix(path, DefaultPath) {
		return strings.TrimSuffix(path, DefaultPath)
	}
	return "/"
}

type APIValidator struct {
	router routers.Router
	next   http.Handler
	log    *logger.UPPLogger
}

type ValidatorConfig struct {
	Filename string
	AppName  string
	AppPort  string
	Log      *logger.UPPLogger
}

func NewAPIValidator(cfg ValidatorConfig, next http.Handler) (*APIValidator, error) {
	doc, err := openapi3.NewLoader().LoadFromFile(cfg.Filename)
	if err != nil {
		return nil, err
	}
	doc.AddServer(&openapi3.Server{
		ExtensionProps: openapi3.ExtensionProps{},
		URL:            "http://localhost:8080",
		Description:    "",
		Variables:      nil,
	})
	doc.AddServer(&openapi3.Server{
		ExtensionProps: openapi3.ExtensionProps{},
		URL:            "http://" + cfg.AppName + ":" + cfg.AppPort,
		Description:    "",
		Variables:      nil,
	})
	err = doc.Validate(context.Background())
	if err != nil {
		return nil, err
	}
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, err
	}
	return &APIValidator{
		router: router,
		next:   next,
		log:    cfg.Log,
	}, nil
}

func (v *APIValidator) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	route, pathParams, err := v.router.FindRoute(req)
	if err != nil {
		v.log.WithError(err).Error("failed to find route")
		return
	}
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:     req,
		PathParams:  pathParams,
		QueryParams: req.URL.Query(),
		Route:       route,
		Options: &openapi3filter.Options{
			MultiError:         true,
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
	}
	err = openapi3filter.ValidateRequest(ctx, requestValidationInput)
	if err != nil {
		v.log.WithError(err).Error("failed to validate request")
	}
	v.next.ServeHTTP(w, req)
}
