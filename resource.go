package api

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/Financial-Times/service-status-go/buildinfo"

	yaml "gopkg.in/yaml.v2"
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
	if strings.HasSuffix(path, "/__api") {
		return strings.TrimSuffix(path, "/__api")
	}
	return "/"
}
