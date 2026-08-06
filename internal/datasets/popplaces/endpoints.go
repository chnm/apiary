package popplaces

import "github.com/chnm/apiary/internal/httpx"

func Endpoints(baseURL string) []httpx.Endpoint {
	return []httpx.Endpoint{
		{Name: "Populated places: A list of counties in a state", URL: baseURL + "/pop-places/state/ma/county/"},
		{Name: "Populated places: A list of places in a county", URL: baseURL + "/pop-places/county/cas_ventura/place/"},
		{Name: "Populated places: Information about a populated place", URL: baseURL + "/pop-places/place/611119/"},
	}
}
