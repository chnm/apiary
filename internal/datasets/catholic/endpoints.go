package catholic

import "github.com/chnm/apiary/internal/httpx"

func Endpoints(baseURL string) []httpx.Endpoint {
	return []httpx.Endpoint{
		{Name: "Roman Catholic Dioceses in North America", URL: baseURL + "/catholic-dioceses/"},
		{Name: "Roman Catholic Dioceses in North America: number established per decade", URL: baseURL + "/catholic-dioceses/per-decade/"},
	}
}
