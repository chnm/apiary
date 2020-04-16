package dataapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Endpoint describes an endpoint available in this API and provides a sample URL.
type Endpoint struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// EndpointsHandler describes the endpoints that are available in this API, with
// sample URLs to show how the API works.
func (s *Server) EndpointsHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// These endpoints should correspond to the routes
		endpoints := []Endpoint{
			{"Historial U.S. county boundaries by date from the Atlas of Historical County Boundaries",
				"http://" + r.Host + "/ahcb/counties/1844-05-08/"},
			{"Historial U.S. county boundaries by date and county ID from the Atlas of Historical County Boundaries",
				"http://" + r.Host + "/ahcb/counties/1844-05-08/id/mas_essex,mas_middlesex/"},
			{"Historial U.S. county boundaries by date and state code from the Atlas of Historical County Boundaries",
				"http://" + r.Host + "/ahcb/counties/1844-05-08/state-code/nh,vt/"},
			{"Historial U.S. state boundaries by date from the Atlas of Historical County Boundaries",
				"http://" + r.Host + "/ahcb/states/1820-05-10/"},
			{"Roman Catholic Dioceses in North America",
				"http://" + r.Host + "/catholic-dioceses/"},
			{"Countries in North America from Natural Earth",
				"http://" + r.Host + "/ne/northamerica/"},
			{"Presbyterian statistics, 1826-1926",
				"http://" + r.Host + "/presbyterians/"},
		}

		response, _ := json.MarshalIndent(endpoints, "", "  ")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(response))
	}
}
