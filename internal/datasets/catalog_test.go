package datasets_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/chnm/apiary/internal/datasets/ahcb"
	"github.com/chnm/apiary/internal/datasets/apb"
	"github.com/chnm/apiary/internal/datasets/bom"
	"github.com/chnm/apiary/internal/datasets/catholic"
	"github.com/chnm/apiary/internal/datasets/naturalearth"
	"github.com/chnm/apiary/internal/datasets/pinkertons"
	"github.com/chnm/apiary/internal/datasets/popplaces"
	"github.com/chnm/apiary/internal/datasets/presbyterians"
	"github.com/chnm/apiary/internal/datasets/relcensus"
	"github.com/chnm/apiary/internal/httpx"
	"github.com/gorilla/mux"
)

func TestDatasetCatalogsMatchRegisteredRoutes(t *testing.T) {
	const baseURL = "https://data.example"
	tests := []struct {
		name      string
		wantCount int
		register  func(*mux.Router)
		endpoints []httpx.Endpoint
	}{
		{name: "AHCB", wantCount: 5, register: ahcb.New(nil).RegisterRoutes, endpoints: ahcb.Endpoints(baseURL)},
		{name: "APB", wantCount: 11, register: apb.New(nil).RegisterRoutes, endpoints: apb.Endpoints(baseURL)},
		{name: "BOM", wantCount: 9, register: bom.New(nil).RegisterRoutes, endpoints: bom.Endpoints(baseURL)},
		{name: "Catholic", wantCount: 2, register: catholic.New(nil).RegisterRoutes, endpoints: catholic.Endpoints(baseURL)},
		{name: "Natural Earth", wantCount: 1, register: naturalearth.New(nil).RegisterRoutes, endpoints: naturalearth.Endpoints(baseURL)},
		{name: "populated places", wantCount: 3, register: popplaces.New(nil).RegisterRoutes, endpoints: popplaces.Endpoints(baseURL)},
		{name: "Presbyterians", wantCount: 1, register: presbyterians.New(nil).RegisterRoutes, endpoints: presbyterians.Endpoints(baseURL)},
		{name: "religious census", wantCount: 4, register: relcensus.New(nil).RegisterRoutes, endpoints: relcensus.Endpoints(baseURL)},
		{name: "Pinkertons", wantCount: 5, register: pinkertons.New(nil).RegisterRoutes, endpoints: pinkertons.Endpoints(baseURL)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.endpoints) != tt.wantCount {
				t.Fatalf("endpoint count = %d, want %d", len(tt.endpoints), tt.wantCount)
			}
			router := mux.NewRouter()
			tt.register(router)
			for _, endpoint := range tt.endpoints {
				assertCatalogRoute(t, router, endpoint.URL)
				for _, example := range endpoint.Examples {
					assertCatalogRoute(t, router, example.URL)
				}
			}
		})
	}
}

func assertCatalogRoute(t *testing.T, router *mux.Router, endpointURL string) {
	t.Helper()
	parsed, err := url.Parse(endpointURL)
	if err != nil {
		t.Fatalf("parse endpoint URL %q: %v", endpointURL, err)
	}
	request := httptest.NewRequest(http.MethodGet, parsed.RequestURI(), nil)
	if matched := router.Match(request, &mux.RouteMatch{}); !matched {
		t.Errorf("catalog URL %q does not match a registered route", endpointURL)
	}
}
