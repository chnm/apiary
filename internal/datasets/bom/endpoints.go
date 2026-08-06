package bom

import "github.com/chnm/apiary/internal/httpx"

// Endpoints returns the BOM entries for the API's root endpoint catalog.
func Endpoints(baseURL string) []httpx.Endpoint {
	return []httpx.Endpoint{
		{
			Name: "BOM: Total records",
			URL:  baseURL + "/bom/totalbills?type=weekly",
			Examples: []httpx.ExampleURL{
				{URL: baseURL + "/bom/totalbills?type=causes", Purpose: "Total records for causes of death"},
				{URL: baseURL + "/bom/totalbills?type=christenings", Purpose: "Total records for christenings"},
				{URL: baseURL + "/bom/totalbills?type=weekly", Purpose: "Total records for weekly bills"},
				{URL: baseURL + "/bom/totalbills?type=general", Purpose: "Total records for general bills"},
			},
		},
		{
			Name: "BOM: Parishes",
			URL:  baseURL + "/bom/parishes",
		},
		{
			Name: "BOM: Completion Statistics",
			URL:  baseURL + "/bom/statistics",
			Examples: []httpx.ExampleURL{
				{URL: baseURL + "/bom/statistics?type=weekly", Purpose: "Group completed bill transcriptions by week"},
				{URL: baseURL + "/bom/statistics?type=yearly", Purpose: "Group completed bill transcriptions by year"},
				{URL: baseURL + "/bom/statistics?type=parish-yearly", Purpose: "Group weekly bills by parish and year"},
				{URL: baseURL + "/bom/statistics?type=parish-yearly&parish=All%20Hallows%20Barking", Purpose: "Group weekly bills by parish and year"},
			},
		},
		{
			Name: "BOM: Bills data with parish polygons",
			URL:  baseURL + "/bom/shapefiles",
			Examples: []httpx.ExampleURL{
				{URL: baseURL + "/bom/shapefiles?year=1665", Purpose: "Bills data with parish polygons for a specific year"},
				{URL: baseURL + "/bom/shapefiles?start-year=1664&end-year=1666", Purpose: "Bills data with parish polygons for a range of years"},
				{URL: baseURL + "/bom/shapefiles?start-year=1664&end-year=1666&bill-type=weekly&count-type=buried", Purpose: "Bills data with parish polygons filtered by bill type and count type"},
			},
		},
		{
			Name: "BOM: Bills of Mortality",
			URL:  baseURL + "/bom/bills?start-year=1636&end-year=1754",
			Examples: []httpx.ExampleURL{
				{URL: baseURL + "/bom/bills?start-year=1636&end-year=1754&bill-type=weekly&parish=1,3,17,28&limit=50&offset=0", Purpose: "Weekly bills for a specific parish or set of parishes by ID. Bill type can be: 'weekly' or 'general'."},
				{URL: baseURL + "/bom/bills?start-year=1636&end-year=1754&count-type=buried&limit=50&offset=0", Purpose: "Bills data for a specific count type (buried or plague). Specific parishes can be provided."},
				{URL: baseURL + "/bom/bills?start-year=1636&end-year=1754&bill-type=weekly&count-type=buried&limit=50&offset=0", Purpose: "Bill type (weekly or general) and count type (buried or plague) can be specific. Specific parishes can be provided."},
				{URL: baseURL + "/bom/bills?start-year=1665&end-year=1665&start-week=10&end-week=15&limit=50&offset=0", Purpose: "Filter bills by week number range (1-53). Useful for seasonal analysis or specific time periods within a year."},
				{URL: baseURL + "/bom/bills?start-year=1665&end-year=1665&start-week=50&bill-type=weekly&count-type=plague&limit=50&offset=0", Purpose: "Combine week number filtering with other parameters. Example shows plague deaths from week 50 onwards in 1665."},
				{URL: baseURL + "/bom/bills?start-year=1636&end-year=1754&missing=false&illegible=false&limit=50&offset=0", Purpose: "Filter out missing and illegible records. Parameters accept true/false values."},
				{URL: baseURL + "/bom/bills?start-year=1636&end-year=1754&missing=true&limit=50&offset=0", Purpose: "Show only missing records. Can be combined with other filtering parameters."},
			},
		},
		{
			Name: "BOM: Causes of Death",
			URL:  baseURL + "/bom/causes?start-year=1648&end-year=1754&limit=50&offset=0",
			Examples: []httpx.ExampleURL{
				{URL: baseURL + "/bom/causes", Purpose: "Return all causes of death with bill_type indicating 'weekly' or 'general' bills"},
				{URL: baseURL + "/bom/causes?start-year=1648&end-year=1754", Purpose: "Causes of death for a specific year range with bill_type parameter"},
				{URL: baseURL + "/bom/causes?start-year=1648&end-year=1754&bill-type=general&id=aged,drowned", Purpose: "Causes of death for a specific year range and cause IDs with bill_type parameter"},
			},
		},
		{
			Name: "BOM: Christenings",
			URL:  baseURL + "/bom/christenings?start-year=1669&end-year=1754&limit=50&offset=0",
			Examples: []httpx.ExampleURL{
				{URL: baseURL + "/bom/christenings?start-year=1669&end-year=1754&id=1,3,17,28", Purpose: "Christenings for a specific year range and parish IDs"},
				{URL: baseURL + "/bom/christenings?start-year=1669&end-year=1754&bill-type=weekly", Purpose: "Christenings for a specific year range from weekly bills"},
			},
		},
		{
			Name: "BOM: List of unique Causes of Death",
			URL:  baseURL + "/bom/list-deaths",
		},
		{
			Name: "BOM: List of unique Christening Parishes",
			URL:  baseURL + "/bom/list-christenings",
		},
	}
}
