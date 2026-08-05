package presbyterians

import "github.com/chnm/apiary/internal/httpx"

func Endpoints(baseURL string) []httpx.Endpoint {
	return []httpx.Endpoint{{Name: "Presbyterian statistics, 1826-1926", URL: baseURL + "/presbyterians/"}}
}
