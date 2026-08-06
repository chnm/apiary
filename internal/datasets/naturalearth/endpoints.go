package naturalearth

import "github.com/chnm/apiary/internal/httpx"

func Endpoints(baseURL string) []httpx.Endpoint {
	return []httpx.Endpoint{{
		Name: "Countries from Natural Earth",
		URL:  baseURL + "/ne/globe?location=Europe",
		Examples: []httpx.ExampleURL{
			{URL: baseURL + "/ne/globe", Purpose: "All available polygons for all countries"},
			{URL: baseURL + "/ne/globe?location=Europe", Purpose: "All available polygons for Europe"},
			{URL: baseURL + "/ne/globe?location=Europe&location=Asia", Purpose: "All available polygons for Europe and Asia"},
		},
	}}
}
