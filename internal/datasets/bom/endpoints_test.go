package bom

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
)

func TestEndpointCatalogMatchesRegisteredRoutes(t *testing.T) {
	router := mux.NewRouter()
	New(nil).RegisterRoutes(router)

	endpoints := Endpoints("https://data.example")
	if len(endpoints) != 9 {
		t.Fatalf("endpoint count = %d, want 9", len(endpoints))
	}

	for _, endpoint := range endpoints {
		assertRouteRegistered(t, router, endpoint.URL)
		for _, example := range endpoint.Examples {
			assertRouteRegistered(t, router, example.URL)
		}
	}
}

func assertRouteRegistered(t *testing.T, router *mux.Router, endpointURL string) {
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
