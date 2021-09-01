package dataapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

		proto := "http://"
		if r.TLS != nil {
			proto = "https://"
		}
		baseurl := proto + r.Host

		// These endpoints should correspond to the routes
		endpoints := []Endpoint{
			{"Historial U.S. county boundaries by date from the Atlas of Historical County Boundaries",
				baseurl + "/ahcb/counties/1844-05-08/"},
			{"Historial U.S. county boundaries by date and county ID from the Atlas of Historical County Boundaries",
				baseurl + "/ahcb/counties/1844-05-08/id/mas_essex,mas_middlesex/"},
			{"Historial U.S. county boundaries by date and state/territory ID from the Atlas of Historical County Boundaries",
				baseurl + "/ahcb/counties/1834-05-08/state-terr-id/nc_state,sc_state/"},
			{"Historial U.S. county boundaries by date and state code from the Atlas of Historical County Boundaries",
				baseurl + "/ahcb/counties/1844-05-08/state-code/nh,vt/"},
			{"Historial U.S. state boundaries by date from the Atlas of Historical County Boundaries",
				baseurl + "/ahcb/states/1820-05-10/"},
			{"APB: Featured verses",
				baseurl + "/apb/index/featured/"},
			{"APB: Top verses",
				baseurl + "/apb/index/top/"},
			{"APB: Verses in biblical order",
				baseurl + "/apb/index/biblical/"},
			{"APB: Verses in chronological order of peak quotations",
				baseurl + "/apb/index/peaks/"},
			{"APB: Verse",
				baseurl + "/apb/verse?ref=Luke+18:16"},
			{"APB: Verse trend",
				baseurl + "/apb/verse-trend?ref=Luke+18:16&corpus=chronam"},
			{"APB: Verse quotations",
				baseurl + "/apb/verse-quotations?ref=Luke+18:16"},
			{"APB: Bible trend",
				baseurl + "/apb/bible-trend/"},
			{"APB: Bible similarity",
				baseurl + "/apb/bible-similarity/"},
			{"Roman Catholic Dioceses in North America",
				baseurl + "/catholic-dioceses/"},
			{"Roman Catholic Dioceses in North America: number established per decade",
				baseurl + "/catholic-dioceses/per-decade/"},
			{"Countries in North America from Natural Earth",
				baseurl + "/ne/northamerica/"},
			{"Populated places: A list of counties in a state",
				baseurl + "/pop-places/state/ma/county/"},
			{"Populated places: A list of places in a county",
				baseurl + "/pop-places/county/cas_ventura/place/"},
			{"Populated places: Information about a populated place",
				baseurl + "/pop-places/place/611119/"},
			{"Presbyterian statistics, 1826-1926",
				baseurl + "/presbyterians/"},
			{"Religious Bodies Census denomination families",
				baseurl + "/relcensus/denomination-families"},
			{"Religious Bodies Census denominations",
				baseurl + "/relcensus/denominations"},
			{"Religious Bodies Census membership data for a denomination in a city in a year",
				baseurl + "/relcensus/city-membership?year=1926&denomination=Protestant+Episcopal+Church"},
			{"Religious Bodies Census membership data aggregated for all denominations in a city in a year",
				baseurl + "/relcensus/city-total-membership?year=1926"},
		}

		response, _ := json.MarshalIndent(endpoints, "", "  ")
		resp := strings.Replace(string(response), "\\u0026", "&", -1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, resp)
	}
}
