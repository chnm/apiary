package dataapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Endpoint describes an endpoint available in this API and provides a sample path.
type Endpoint struct {
	Name string `json:"name"`
	URL  string `json:"path"`
}

// EndpointsHandler describes the endpoints that are available in this API, with
// sample URLs to show how the API works.
func (s *Server) EndpointsHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// These endpoints should correspond to the routes
		endpoints := []Endpoint{
			{"Historial U.S. county boundaries by date from the Atlas of Historical County Boundaries",
				"/ahcb/counties/1844-05-08/"},
			{"Historial U.S. county boundaries by date and county ID from the Atlas of Historical County Boundaries",
				"/ahcb/counties/1844-05-08/id/mas_essex,mas_middlesex/"},
			{"Historial U.S. county boundaries by date and state/territory ID from the Atlas of Historical County Boundaries",
				"/ahcb/counties/1834-05-08/state-terr-id/nc_state,sc_state/"},
			{"Historial U.S. county boundaries by date and state code from the Atlas of Historical County Boundaries",
				"/ahcb/counties/1844-05-08/state-code/nh,vt/"},
			{"Historial U.S. state boundaries by date from the Atlas of Historical County Boundaries",
				"/ahcb/states/1820-05-10/"},
			{"APB: Featured verses",
				"/apb/index/featured/"},
			{"APB: Verse",
				"/apb/verse?ref=Luke+18:16"},
			{"APB: Verse trend",
				"/apb/verse-trend?ref=Luke+18:16&corpus=chronam"},
			{"APB: Verse quotations",
				"/apb/verse-quotations?ref=Luke+18:16"},
			{"Roman Catholic Dioceses in North America",
				"/catholic-dioceses/"},
			{"Roman Catholic Dioceses in North America: number established per decade",
				"/catholic-dioceses/per-decade/"},
			{"Countries in North America from Natural Earth",
				"/ne/northamerica/"},
			{"Populated places: A list of counties in a state",
				"/pop-places/state/ma/county/"},
			{"Populated places: A list of places in a county",
				"/pop-places/county/cas_ventura/place/"},
			{"Populated places: Information about a populated place",
				"/pop-places/place/611119/"},
			{"Presbyterian statistics, 1826-1926",
				"/presbyterians/"},
		}

		response, _ := json.MarshalIndent(endpoints, "", "  ")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(response))
	}
}
