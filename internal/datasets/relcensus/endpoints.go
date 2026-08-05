package relcensus

import "github.com/chnm/apiary/internal/httpx"

func Endpoints(baseURL string) []httpx.Endpoint {
	return []httpx.Endpoint{
		{Name: "Religious Bodies Census denomination families", URL: baseURL + "/relcensus/denomination-families"},
		{Name: "Religious Bodies Census denominations", URL: baseURL + "/relcensus/denominations"},
		{Name: "Religious Bodies list of all cities", URL: baseURL + "/relcensus/cities"},
		{
			Name: "Religious Bodies Census membership data for a denomination in a city in a year",
			URL:  baseURL + "/relcensus/city-membership?year=1926&denomination=Protestant+Episcopal+Church",
			Examples: []httpx.ExampleURL{
				{URL: baseURL + "/relcensus/city-membership?year=1926&denomination=Church+of+God+in+Christ", Purpose: "Membership data for a specific denomination in each city"},
				{URL: baseURL + "/relcensus/city-membership?year=1926&denominationFamily=Pentecostal", Purpose: "Membership data aggregated for a denomination family in each city"},
				{URL: baseURL + "/relcensus/city-membership?year=1926", Purpose: "Membership data aggregated for all denominations in each city"},
			},
		},
	}
}
