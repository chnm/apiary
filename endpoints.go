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
			{
				"Historial U.S. county boundaries by date from the Atlas of Historical County Boundaries",
				baseurl + "/ahcb/counties/1844-05-08/",
				nil,
			},
			{
				"Historial U.S. county boundaries by date and county ID from the Atlas of Historical County Boundaries",
				baseurl + "/ahcb/counties/1844-05-08/id/mas_essex,mas_middlesex/",
				nil,
			},
			{
				"Historial U.S. county boundaries by date and state/territory ID from the Atlas of Historical County Boundaries",
				baseurl + "/ahcb/counties/1834-05-08/state-terr-id/nc_state,sc_state/",
				nil,
			},
			{
				"Historial U.S. county boundaries by date and state code from the Atlas of Historical County Boundaries",
				baseurl + "/ahcb/counties/1844-05-08/state-code/nh,vt/",
				nil,
			},
			{
				"Historial U.S. state boundaries by date from the Atlas of Historical County Boundaries",
				baseurl + "/ahcb/states/1820-05-10/",
				nil,
			},
			{
				"APB: Featured verses",
				baseurl + "/apb/index/featured",
				nil,
			},
			{
				"APB: Top verses",
				baseurl + "/apb/index/top",
				nil,
			},
			{
				"APB: Verses in biblical order",
				baseurl + "/apb/index/biblical",
				nil,
			},
			{
				"APB: Verses in chronological order of peak quotations",
				baseurl + "/apb/index/peaks",
				nil,
			},
			{
				"APB: All verses in biblical order",
				baseurl + "/apb/index/all",
				nil,
			},
			{
				"APB: Verse",
				baseurl + "/apb/verse?ref=Luke+18:16",
				nil,
			},
			{
				"APB: Verse trend",
				baseurl + "/apb/verse-trend?ref=Luke+18:16&corpus=chronam",
				nil,
			},
			{
				"APB: Verse quotations",
				baseurl + "/apb/verse-quotations?ref=Luke+18:16",
				nil,
			},
			{
				"APB: Bible trend",
				baseurl + "/apb/bible-trend",
				nil,
			},
			{
				"APB: Bible similarity",
				baseurl + "/apb/bible-similarity",
				nil,
			},
			{
				"APB: Books of the Bible",
				baseurl + "/apb/bible-books",
				nil,
			},
			{
				"BOM: Total records",
				baseurl + "/bom/totalbills?type=Weekly",
				[]ExampleURL{
					{
						baseurl + "/bom/totalbills?type=Causes",
						"Total records for causes of death",
					},
					{
						baseurl + "/bom/totalbills?type=Christenings",
						"Total records for christenings",
					},
					{
						baseurl + "/bom/totalbills?type=Weekly",
						"Total records for weekly bills",
					},
					{
						baseurl + "/bom/totalbills?type=General",
						"Total records for general bills",
					},
				},
			},
			{
				"BOM: Parishes",
				baseurl + "/bom/parishes",
				nil,
			},
			{
				"BOM: Parish polygons",
				baseurl + "/bom/geometries",
				nil,
			},
			{
				"BOM: Bills of Mortality",
				baseurl + "/bom/bills?start-year=1636&end-year=1754",
				[]ExampleURL{
					{
						baseurl + "/bom/bills?start-year=1636&end-year=1754&bill-type=Weekly&parish=1,3,17,28&limit=50&offset=0",
						"Weekly bills for a specific parish or set of parishes by ID. Bill type can be: Weekly or General.",
					},
					{
						baseurl + "/bom/bills?start-year=1636&end-year=1754&count-type=Buried&limit=50&offset=0",
						"Bills data for a specific count type (Buried or Plague). Specific parishes can be provided.",
					},
					{
						baseurl + "/bom/bills?start-year=1636&end-year=1754&bill-type=Weekly&count-type=Buried&limit=50&offset=0",
						"Bill type (Weekly or General) and count type (Buried or Plague) can be specific. Specific parishes can be provided.",
					},
				},
			},
			{
				"BOM: Causes of Death",
				baseurl + "/bom/causes?start-year=1648&end-year=1754&limit=50&offset=0",
				[]ExampleURL{
					{
						baseurl + "/bom/causes",
						"Return all causes of death",
					},
					{
						baseurl + "/bom/causes?start-year=1648&end-year=1754",
						"Causes of death for a specific year range",
					},
					{
						baseurl + "/bom/causes?start-year=1648&end-year=1754&id=Aged,Drowned",
						"Causes of death for a specific year range and cause IDs",
					},
				},
			},
			{
				"BOM: Christenings",
				baseurl + "/bom/christenings?start-year=1669&end-year=1754&limit=50&offset=0",
				[]ExampleURL{
					{
						baseurl + "/bom/christenings?start-year=1669&end-year=1754&id=1,3,17,28",
						"Christenings for a specific year range and parish IDs",
					},
				},
			},
			{
				"BOM: List of unique Causes of Death",
				baseurl + "/bom/list-deaths",
				nil,
			},
			{
				"BOM: List of unique Christening Parishes",
				baseurl + "/bom/list-christenings",
				nil,
			},
			{
				"Roman Catholic Dioceses in North America",
				baseurl + "/catholic-dioceses/",
				nil,
			},
			{
				"Roman Catholic Dioceses in North America: number established per decade",
				baseurl + "/catholic-dioceses/per-decade/",
				nil,
			},
			{
				"Countries from Natural Earth",
				baseurl + "/ne/globe?location=Europe",
				[]ExampleURL{
					{
						baseurl + "/ne/globe",
						"All available polygons for all countries",
					},
					{
						baseurl + "/ne/globe?location=Europe",
						"All available polygons for Europe",
					},
					{
						baseurl + "/ne/globe?location=Europe&location=Asia",
						"All available polygons for Europe and Asia",
					},
				},
			},
			{
				"Populated places: A list of counties in a state",
				baseurl + "/pop-places/state/ma/county/",
				nil,
			},
			{
				"Populated places: A list of places in a county",
				baseurl + "/pop-places/county/cas_ventura/place/",
				nil,
			},
			{
				"Populated places: Information about a populated place",
				baseurl + "/pop-places/place/611119/",
				nil,
			},
			{
				"Presbyterian statistics, 1826-1926",
				baseurl + "/presbyterians/",
				nil,
			},
			{
				"Religious Bodies Census denomination families",
				baseurl + "/relcensus/denomination-families",
				nil,
			},
			{
				"Religious Bodies Census denominations",
				baseurl + "/relcensus/denominations",
				nil,
			},
			{
				"Religious Bodies Census membership data for a denomination in a city in a year",
				baseurl + "/relcensus/city-membership?year=1926&denomination=Protestant+Episcopal+Church",
				[]ExampleURL{
					{
						baseurl + "/relcensus/city-membership?year=1926&denomination=Church+of+God+in+Christ",
						"Membership data for a specific denomination in each city",
					},
					{
						baseurl + "/relcensus/city-membership?year=1926&denominationFamily=Pentecostal",
						"Membership data aggregated for a denomination family in each city",
					},
					{
						baseurl + "/relcensus/city-membership?year=1926",
						"Membership data aggregated for all denominations in each city",
					},
				},
			},
			{
				"Cache test",
				baseurl + "/cache",
				[]ExampleURL{
					{baseurl + "/cache", "Cached response"},
					{baseurl + "/cache?nocache", "Uncached response"},
				},
			},
		}

		response, _ := json.MarshalIndent(endpoints, "", "  ")
		resp := strings.Replace(string(response), "\\u0026", "&", -1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, resp)
	}
}
