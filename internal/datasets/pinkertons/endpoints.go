package pinkertons

import "github.com/chnm/apiary/internal/httpx"

func Endpoints(baseURL string) []httpx.Endpoint {
	return []httpx.Endpoint{
		{
			Name: "Pinkertons: Activities (first 500 by default)",
			URL:  baseURL + "/pinkertons/activities",
			Examples: []httpx.ExampleURL{
				{URL: baseURL + "/pinkertons/activities?limit=10", Purpose: "First 10 activities with location coordinates"},
				{URL: baseURL + "/pinkertons/activities?limit=10&offset=10", Purpose: "Next 10 activities with location coordinates"},
				{URL: baseURL + "/pinkertons/activities?operative=John+Doe", Purpose: "Follow a specific operative"},
				{URL: baseURL + "/pinkertons/activities?subject=Jane+Smith", Purpose: "Activities related to a specific subject"},
				{URL: baseURL + "/pinkertons/activities?start_date=1900-01-01&end_date=1900-12-31", Purpose: "Activities within a date range"},
				{URL: baseURL + "/pinkertons/activities?limit=50&start_date=1900-01-01", Purpose: "First 50 activities from 1900 onwards"},
			},
		},
		{Name: "Pinkertons: Activity by ID with locations", URL: baseURL + "/pinkertons/activities/1"},
		{Name: "Pinkertons: All locations with coordinates", URL: baseURL + "/pinkertons/locations"},
		{Name: "Pinkertons: List of unique operatives", URL: baseURL + "/pinkertons/operatives"},
		{Name: "Pinkertons: List of unique subjects", URL: baseURL + "/pinkertons/subjects"},
	}
}
