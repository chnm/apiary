package apiary

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
				baseurl + "/ahcb/counties/1844-05-08/",
				nil,
			},
			{"Historial U.S. county boundaries by date and county ID from the Atlas of Historical County Boundaries",
				baseurl + "/ahcb/counties/1844-05-08/id/mas_essex,mas_middlesex/",
				nil},
			{"Historial U.S. county boundaries by date and state/territory ID from the Atlas of Historical County Boundaries",
				baseurl + "/ahcb/counties/1834-05-08/state-terr-id/nc_state,sc_state/",
				nil},
			{"Historial U.S. county boundaries by date and state code from the Atlas of Historical County Boundaries",
				baseurl + "/ahcb/counties/1844-05-08/state-code/nh,vt/",
				nil},
			{"Historial U.S. state boundaries by date from the Atlas of Historical County Boundaries",
				baseurl + "/ahcb/states/1820-05-10/",
				nil},
			{"APB: Featured verses",
				baseurl + "/apb/index/featured",
				nil},
			{"APB: Top verses",
				baseurl + "/apb/index/top",
				nil},
			{"APB: Verses in biblical order",
				baseurl + "/apb/index/biblical",
				nil},
			{"APB: Verses in chronological order of peak quotations",
				baseurl + "/apb/index/peaks",
				nil},
			{"APB: All verses in biblical order",
				baseurl + "/apb/index/all",
				nil},
			{"APB: Verse",
				baseurl + "/apb/verse?ref=Luke+18:16",
				nil},
			{"APB: Verse trend",
				baseurl + "/apb/verse-trend?ref=Luke+18:16&corpus=chronam",
				nil},
			{"APB: Verse quotations",
				baseurl + "/apb/verse-quotations?ref=Luke+18:16",
				nil},
			{"APB: Bible trend",
				baseurl + "/apb/bible-trend",
				nil},
			{"APB: Bible similarity",
				baseurl + "/apb/bible-similarity",
				nil},
			{"APB: Books of the Bible",
				baseurl + "/apb/bible-books",
				nil},
			{"BOM: Parishes",
				baseurl + "/bom/parishes",
				nil},
			{"BOM: Bills of Mortality",
				baseurl + "/bom/bills?startYear=1669&endYear=1754",
				nil},
			{"BOM: General Bills of Mortality",
				baseurl + "/bom/generalbills?startYear=1669&endYear=1754",
				nil},
			{"BOM: Causes of Death",
				baseurl + "/bom/causes",
				nil},
			{"BOM: Christenings",
				baseurl + "/bom/christenings?startYear=1669&endYear=1754",
				nil},
			{"Roman Catholic Dioceses in North America",
				baseurl + "/catholic-dioceses/",
				nil},
			{"Roman Catholic Dioceses in North America: number established per decade",
				baseurl + "/catholic-dioceses/per-decade/",
				nil},
			{"Countries in North America from Natural Earth",
				baseurl + "/ne/northamerica/",
				nil},
			{"Populated places: A list of counties in a state",
				baseurl + "/pop-places/state/ma/county/",
				nil},
			{"Populated places: A list of places in a county",
				baseurl + "/pop-places/county/cas_ventura/place/",
				nil},
			{"Populated places: Information about a populated place",
				baseurl + "/pop-places/place/611119/",
				nil},
			{"Presbyterian statistics, 1826-1926",
				baseurl + "/presbyterians/",
				nil},
			{"Religious Bodies Census denomination families",
				baseurl + "/relcensus/denomination-families",
				nil},
			{"Religious Bodies Census denominations",
				baseurl + "/relcensus/denominations",
				nil},
			{"Religious Bodies Census membership data for a denomination in a city in a year",
				baseurl + "/relcensus/city-membership?year=1926&denomination=Protestant+Episcopal+Church",
				[]ExampleURL{
					{baseurl + "/relcensus/city-membership?year=1926&denomination=Church+of+God+in+Christ",
						"Membership data for a specific denomination in each city"},
					{baseurl + "/relcensus/city-membership?year=1926&denominationFamily=Pentecostal",
						"Membership data aggregated for a denomination family in each city"},
					{baseurl + "/relcensus/city-membership?year=1926",
						"Membership data aggregated for all denominations in each city"},
				}},
		}

		response, _ := json.MarshalIndent(endpoints, "", "  ")
		resp := strings.Replace(string(response), "\\u0026", "&", -1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, resp)
	}
}

// []ExampleURL{{baseurl + "/ahcb/counties/1844-05-08/", "County on a specific date"}},
