package api

import (
	"net/http/httptest"
	"testing"

	yaml "gopkg.in/yaml.v2"

	"github.com/Financial-Times/service-status-go/buildinfo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPI(t *testing.T) {
	req := httptest.NewRequest("GET", "/__api", nil)
	req.Header.Add("X-Original-Request-URL", "https://pub-dynpub-uk-up.ft.com/__publish-carousel/__api")

	actual := httptest.NewRecorder()
	endpoint, err := NewAPIEndpointForYAML([]byte(apiExample))
	assert.NoError(t, err)

	endpoint.ServeHTTP(actual, req)

	yml := make(map[string]interface{})
	err = yaml.Unmarshal(actual.Body.Bytes(), &yml)
	require.NoError(t, err)

	assert.Equal(t, "pub-dynpub-uk-up.ft.com", yml["host"])
	assert.Equal(t, "/__publish-carousel", yml["basePath"])
	assert.Equal(t, []interface{}{"https"}, yml["schemes"])
}

func TestAPINoOriginalRequestURL(t *testing.T) {
	req := httptest.NewRequest("GET", "/__api", nil)

	actual := httptest.NewRecorder()
	endpoint, err := NewAPIEndpointForYAML([]byte(apiExample))
	assert.NoError(t, err)

	endpoint.ServeHTTP(actual, req)

	assert.Equal(t, []byte(apiExample), actual.Body.Bytes())
}

func TestAPIOriginalRequestURLInvalid(t *testing.T) {
	req := httptest.NewRequest("GET", "/__api", nil)
	req.Header.Add("X-Original-Request-URL", ":#")

	actual := httptest.NewRecorder()
	endpoint, err := NewAPIEndpointForYAML([]byte(apiExample))
	assert.NoError(t, err)

	endpoint.ServeHTTP(actual, req)
	assert.Equal(t, apiExample, actual.Body.String())
}

func TestAPIOriginalRequestURLBlank(t *testing.T) {
	req := httptest.NewRequest("GET", "/__api", nil)
	req.Header.Add("X-Original-Request-URL", "   ")

	actual := httptest.NewRecorder()
	endpoint, err := NewAPIEndpointForYAML([]byte(apiExample))
	assert.NoError(t, err)

	endpoint.ServeHTTP(actual, req)
	assert.Equal(t, apiExample, actual.Body.String())
}

func TestAPIOriginalRequestURLUnexpectedPath(t *testing.T) {
	req := httptest.NewRequest("GET", "/__api", nil)
	req.Header.Add("X-Original-Request-URL", "https://pub-dynpub-uk-up.ft.com/__publish-carousel/hello")

	actual := httptest.NewRecorder()
	endpoint, err := NewAPIEndpointForYAML([]byte(apiExample))
	assert.NoError(t, err)

	endpoint.ServeHTTP(actual, req)

	yml := make(map[string]interface{})
	err = yaml.Unmarshal(actual.Body.Bytes(), &yml)
	require.NoError(t, err)

	assert.Equal(t, "pub-dynpub-uk-up.ft.com", yml["host"])
	assert.Equal(t, "/", yml["basePath"])
	assert.Equal(t, []interface{}{"https"}, yml["schemes"])
}

func TestSchemesChangeBasedOnRequestURL(t *testing.T) {
	req := httptest.NewRequest("GET", "/__api", nil)
	req.Header.Add("X-Original-Request-URL", "http://pub-dynpub-uk-up.ft.com/__publish-carousel/hello")

	actual := httptest.NewRecorder()
	endpoint, err := NewAPIEndpointForYAML([]byte(apiExample))
	assert.NoError(t, err)

	endpoint.ServeHTTP(actual, req)

	yml := make(map[string]interface{})
	err = yaml.Unmarshal(actual.Body.Bytes(), &yml)
	require.NoError(t, err)

	assert.Equal(t, "pub-dynpub-uk-up.ft.com", yml["host"])
	assert.Equal(t, "/", yml["basePath"])
	assert.Equal(t, []interface{}{"http"}, yml["schemes"])
}

func TestSchemeDefaultsToHTTPS(t *testing.T) {
	req := httptest.NewRequest("GET", "/__api", nil)
	req.Header.Add("X-Original-Request-URL", "/__publish-carousel/hello")

	actual := httptest.NewRecorder()
	endpoint, err := NewAPIEndpointForYAML([]byte(apiExample))
	assert.NoError(t, err)

	endpoint.ServeHTTP(actual, req)

	yml := make(map[string]interface{})
	err = yaml.Unmarshal(actual.Body.Bytes(), &yml)
	require.NoError(t, err)

	assert.Equal(t, "", yml["host"])
	assert.Equal(t, "/", yml["basePath"])
	assert.Equal(t, []interface{}{"https"}, yml["schemes"])
}

func TestAPIVersionInfo(t *testing.T) {
	req := httptest.NewRequest("GET", "/__api", nil)
	req.Header.Add("X-Original-Request-URL", "https://pub-dynpub-uk-up.ft.com/__publish-carousel/__api")

	actual := httptest.NewRecorder()
	e, err := NewAPIEndpointForYAML([]byte(apiExample))
	assert.NoError(t, err)

	epoint := e.(*endpoint)
	epoint.buildInfo = buildinfo.BuildInfo{Version: "4.0.1"}

	e.ServeHTTP(actual, req)

	yml := make(map[string]interface{})
	err = yaml.Unmarshal(actual.Body.Bytes(), &yml)
	require.NoError(t, err)

	info, ok := yml["info"].(map[interface{}]interface{})
	assert.True(t, ok)
	assert.NotNil(t, info)

	assert.Equal(t, "4.0.1", info["version"])
}

func TestAPIVersionNoInfoSection(t *testing.T) {
	req := httptest.NewRequest("GET", "/__api", nil)
	req.Header.Add("X-Original-Request-URL", "https://pub-dynpub-uk-up.ft.com/__publish-carousel/__api")

	actual := httptest.NewRecorder()
	e, err := NewAPIEndpointForYAML([]byte(shortAPIExample))
	assert.NoError(t, err)

	epoint := e.(*endpoint)
	epoint.buildInfo = buildinfo.BuildInfo{Version: "4.0.1"}

	e.ServeHTTP(actual, req)

	yml := make(map[string]interface{})
	err = yaml.Unmarshal(actual.Body.Bytes(), &yml)
	require.NoError(t, err)

	info, ok := yml["info"].(map[interface{}]interface{})
	assert.False(t, ok)
	assert.Nil(t, info)
}

const apiExample = `swagger: "2.0"

info:
   title: Publish Carousel
   description: A microservice that continuously republishes content and annotations available in the native store.
   version: 0.0.7
   contact:
      name: Dynamic Publishing
      email: Dynamic.Publishing@ft.com

host: api.ft.com

schemes:
   - http
   - https

basePath: /

paths:
   /cycles:
      get:
         summary: Get Cycle Information
         description: Displays state information for all configured cycles.
         tags:
            - Internal API
         responses:
            200:
               description: Shows the state of all configured cycles.
               examples:
                  application/json:
                     -  id: 5118842b62670d2b
                        name: methode-whole-archive
                        type: ThrottledWholeCollection
                        metadata:
                           currentPublishUuid: c372ffba-7a7f-11e6-aca9-d6ece9a77557
                           errors: 0
                           progress: 1
                           state:
                              - stopped
                              - unhealthy
                           completed: 2000
                           total: 100000
                           iteration: 3
                        collection: methode
                        origin: methode-web-pub
                        coolDown: 5m
            500:
               description: An error occurred while processing the cycles into json.
      post:
         summary: Create a new Cycle
         description: Creates and starts a new cycle with the provided configuration.
         tags:
            - Internal API
         consumes:
            - application/json
         parameters:
            -  name: body
               in: body
               required: true
               description: The configuration for the new cycle.
               schema:
                  type: object
                  properties:
                     name:
                        type: string
                     type:
                        type: string
                        enum:
                           - ThrottledWholeCollection
                           - FixedWindow
                           - ScalingWindow
                     origin:
                        type: string
                     collection:
                        type: string
                     coolDown:
                        type: string
                     throttle:
                        type: string
                     timeWindow:
                        type: string
                     minimumThrottle:
                        type: string
                     maximumThrottle:
                        type: string
                  required:
                     - name
                     - type
                     - origin
                     - collection
                     - coolDown
                  example:
                     name: methode-whole-archive-dredd
                     type: ThrottledWholeCollection
                     origin: methode-origin
                     collection: methode
                     coolDown: 5m
                     throttle: 15s
         responses:
            201:
               description: The cycle has been created successfully
            400:
               description: The provided cycle configuration is invalid.
            500:
               description: An error occurred while creating the new cycle, or when adding it to the scheduler.
   /cycles/{id}:
      get:
         summary: Get Cycle Information for ID
         description: Displays state information for the cycle with the given ID
         tags:
            - Internal API
         parameters:
            -  name: id
               in: path
               required: true
               description: The ID of the cycle you would like to view the state for.
               x-example: 5118842b62670d2b
               type: string
         responses:
            200:
               description: Shows the state of the cycle with the provided ID
               examples:
                  application/json:
                     id: 5118842b62670d2b
                     name: methode-whole-archive
                     type: ThrottledWholeCollection
                     metadata:
                        currentPublishUuid: c372ffba-7a7f-11e6-aca9-d6ece9a77557
                        errors: 0
                        progress: 1
                        state:
                           - stopped
                           - unhealthy
                        completed: 2000
                        total: 100000
                        iteration: 3
                     collection: methode
                     origin: methode-web-pub
                     coolDown: 5m
            404:
               description: We couldn't find a cycle with the provided ID.
            500:
               description: An error occurred while processing the cycle into json.
      delete:
         summary: Delete the Cycle
         description: Stops and removes the cycle from the Carousel. Deleted cycles cannot be resumed, and must be recreated.
         tags:
            - Internal API
         parameters:
            -  name: id
               in: path
               required: true
               description: The ID of the cycle you would like to view the state for.
               x-example: 5118842b62670d2b
               type: string
         responses:
            204:
               description: The cycle has been deleted successfully.
            404:
               description: We couldn't find a cycle with the provided ID.
   /cycles/{id}/stop:
      post:
         summary: Stop Cycle
         description: Stops the running of the cycle with the provided ID, and frees up connections to Mongo.
         tags:
            - Internal API
         consumes:
            - application/json
         parameters:
            -  name: id
               in: path
               required: true
               description: The ID of the cycle you would like to view the state for.
               x-example: 7085a0ac743eddd8
               type: string
         responses:
            200:
               description: A stop has been triggered for the cycle.
            404:
               description: We couldn't find a cycle with the provided ID.
   /cycles/{id}/resume:
      post:
         summary: Resume Cycle
         description: Resumes a stopped cycle with ID.
         tags:
            - Internal API
         consumes:
            - application/json
         parameters:
            -  name: id
               in: path
               required: true
               description: The ID of the cycle you would like to view the state for.
               x-example: 7085a0ac743eddd8
               type: string
         responses:
            200:
               description: A resume has been triggered for the cycle.
            404:
               description: We couldn't find a cycle with the provided ID.
   /cycles/{id}/reset:
      post:
         summary: Reset Cycle
         description: Stops the provided cycle, and resets it back to the beginning of its iteration. N.B. the cycle will need to be resumed after being reset.
         tags:
            - Internal API
         consumes:
            - application/json
         parameters:
            -  name: id
               in: path
               required: true
               description: The ID of the cycle you would like to view the state for.
               x-example: 7085a0ac743eddd8
               type: string
         responses:
            200:
               description: A resume has been triggered for the cycle.
            404:
               description: We couldn't find a cycle with the provided ID.
   /scheduler/shutdown:
      post:
         summary: Scheduler Shutdown
         description: Stops all cycles. Useful in a production incident.
         tags:
            - Internal API
         consumes:
            - application/json
         responses:
            200:
               description: Shutdown was successful.
            500:
               description: An error occurred while shutting down the scheduler, please see the logs for details.
   /scheduler/start:
      post:
         summary: Start Scheduler
         description: Resumes all cycles. Useful when the production incident has been resolved.
         tags:
            - Internal API
         consumes:
            - application/json
         responses:
            200:
               description: Shutdown was successful.
            500:
               description: An error occurred while shutting down the scheduler, please see the logs for details.
   /__ping:
      get:
         summary: Ping
         description: Returns "pong" if the server is running.
         produces:
            - text/plain; charset=utf-8
         tags:
            - Health
         responses:
            200:
               description: We return pong in plaintext only.
               examples:
                  text/plain; charset=utf-8: pong
   /__health:
      get:
         summary: Healthchecks
         description: Runs application healthchecks and returns FT Healthcheck style json.
         produces:
            - application/json
         tags:
            - Health
         responses:
            200:
               description: Should always return 200 along with the output of the healthchecks - regardless of whether the healthchecks failed or not. Please inspect the overall ok property to see whether or not the application is healthy.
               examples:
                  application/json:
                     checks:
                        -  businessImpact: "No Business Impact."
                           checkOutput: "OK"
                           lastUpdated: "2017-01-16T10:26:47.222805121Z"
                           name: "UnhealthyCycles"
                           ok: true
                           panicGuide: "https://dewey.ft.com/upp-publish-carousel.html"
                           severity: 1
                           technicalSummary: "At least one of the Carousel cycles is unhealthy. This should be investigated."
                     description: Notifies clients of updates to UPP Lists.
                     name: publish-carousel
                     ok: true
                     schemaVersion: 1
   /__build-info:
      get:
         summary: Build Information
         description: Returns application build info, such as the git repository and revision, the golang version it was built with, and the app release version.
         produces:
            - application/json; charset=UTF-8
         tags:
            - Info
         responses:
            200:
               description: Outputs build information as described in the summary.
               examples:
                  application/json; charset=UTF-8:
                     version: "0.0.7"
                     repository: "https://github.com/Financial-Times/publish-carousel.git"
                     revision: "7cdbdb18b4a518eef3ebb1b545fc124612f9d7cd"
                     builder: "go version go1.6.3 linux/amd64"
                     dateTime: "20161123122615"
   /__gtg:
      get:
         summary: Good To Go
         description: Lightly healthchecks the application, and returns a 200 if it's Good-To-Go.
         tags:
            - Health
         responses:
            200:
               description: The application is healthy enough to perform all its functions correctly - i.e. good to go.
            503:
               description: One or more of the applications healthchecks have failed, so please do not use the app. See the /__health endpoint for more detailed information.`

const shortAPIExample = `swagger: "2.0"

host: api.ft.com

schemes:
   - http
   - https

basePath: /

paths:
   /__build-info:
      get:
         summary: Build Information
         description: Returns application build info, such as the git repository and revision, the golang version it was built with, and the app release version.
         produces:
            - application/json; charset=UTF-8
         tags:
            - Info
         responses:
            200:
               description: Outputs build information as described in the summary.
               examples:
                  application/json; charset=UTF-8:
                     version: "0.0.7"
                     repository: "https://github.com/Financial-Times/publish-carousel.git"
                     revision: "7cdbdb18b4a518eef3ebb1b545fc124612f9d7cd"
                     builder: "go version go1.6.3 linux/amd64"
                     dateTime: "20161123122615"`
