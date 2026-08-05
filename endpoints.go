package apiary

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/chnm/apiary/internal/datasets/ahcb"
	"github.com/chnm/apiary/internal/datasets/apb"
	"github.com/chnm/apiary/internal/datasets/bom"
	"github.com/chnm/apiary/internal/datasets/catholic"
	"github.com/chnm/apiary/internal/datasets/naturalearth"
	"github.com/chnm/apiary/internal/datasets/pinkertons"
	"github.com/chnm/apiary/internal/datasets/popplaces"
	"github.com/chnm/apiary/internal/datasets/presbyterians"
	"github.com/chnm/apiary/internal/datasets/relcensus"
	"github.com/chnm/apiary/internal/httpx"
)

// ExampleURL provides an example URL to a different way of querying the API
// for any given endpoint.
type ExampleURL struct {
	URL     string `json:"url"`
	Purpose string `json:"purpose"`
}

// Endpoint describes an endpoint available in this API and provides a sample path.
type Endpoint struct {
	Name     string       `json:"name"`
	URL      string       `json:"path"`
	Examples []ExampleURL `json:"examples,omitempty"`
}

func appendDatasetEndpoints(endpoints []Endpoint, datasetEndpoints []httpx.Endpoint) []Endpoint {
	for _, datasetEndpoint := range datasetEndpoints {
		var examples []ExampleURL
		if len(datasetEndpoint.Examples) > 0 {
			examples = make([]ExampleURL, len(datasetEndpoint.Examples))
		}
		for i, example := range datasetEndpoint.Examples {
			examples[i] = ExampleURL{
				URL:     example.URL,
				Purpose: example.Purpose,
			}
		}
		endpoints = append(endpoints, Endpoint{
			Name:     datasetEndpoint.Name,
			URL:      datasetEndpoint.URL,
			Examples: examples,
		})
	}
	return endpoints
}

// EndpointsHandler describes the endpoints that are available in this API, with
// sample URLs to show how the API works.
func (s *Server) EndpointsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proto := "http://"
		if r.TLS != nil {
			proto = "https://"
		}
		baseurl := proto + r.Host

		var endpoints []Endpoint
		endpoints = appendDatasetEndpoints(endpoints, ahcb.Endpoints(baseurl))
		endpoints = appendDatasetEndpoints(endpoints, apb.Endpoints(baseurl))
		endpoints = appendDatasetEndpoints(endpoints, bom.Endpoints(baseurl))
		endpoints = appendDatasetEndpoints(endpoints, catholic.Endpoints(baseurl))
		endpoints = appendDatasetEndpoints(endpoints, naturalearth.Endpoints(baseurl))
		endpoints = appendDatasetEndpoints(endpoints, popplaces.Endpoints(baseurl))
		endpoints = appendDatasetEndpoints(endpoints, presbyterians.Endpoints(baseurl))
		endpoints = appendDatasetEndpoints(endpoints, relcensus.Endpoints(baseurl))
		endpoints = appendDatasetEndpoints(endpoints, pinkertons.Endpoints(baseurl))

		response, err := json.MarshalIndent(endpoints, "", "  ")
		if err != nil {
			internalServerError(w, "error marshaling endpoint index", err)
			return
		}
		resp := strings.Replace(string(response), "\\u0026", "&", -1)
		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprint(w, resp); err != nil {
			log.Printf("error writing endpoint index: %v", err)
		}
	}
}
