# API Endpoint

Serves an OpenAPI (Swagger) yml file on a common endpoint, amending values based on the request path.

Currently, the endpoint will amend the following fields if the `X-Original-Request-URL` is set.

* `host` - this is set the host of the `X-Original-Request-URL`
* `schemes` - we assume `https` if the `X-Original-Request-URL` is set
* `basePath` - we take the path of the `X-Original-Request-URL` and trim the api DefaultPath (`/__api`)
* `version` - this will be set to the app version as returned by the `buildInfo` package. Will only be set if the YML contains an `info` field.

The `X-Original-Request-URL` is set as a request header by varnish.

If any errors occur while amending the YML, the endpoint will fall-back to serving the raw `api.yml` file.

# Usage

To use, simply create a new API endpoint type by passing either the file path for the `api.yml`, or a `[]byte` containing the contents of the `api.yml`.

```golang
import  api "github.com/Financial-Times/api-endpoint"
...

apiYml := "./api.yml"

// create the endpoint using the yml file
apiEndpoint, err := api.NewAPIEndpointForFile(apiYml)
if err != nil {
   log.WithError(err).WithField("file", apiYml).Warn("Failed to serve the API Endpoint for this service. Please validate the Swagger YML and the file location.")
}

// or using an []byte
ymlBytes, _ := ioutil.ReadFile(apiYml)
apiEndpoint, err := api.NewAPIEndpointForYAML(ymlBytes)

// serve the endpoint using the default path
router.HandleFunc(api.DefaultPath, apiEndpoint.ServeHTTP)
```
