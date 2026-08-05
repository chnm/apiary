package relcensus

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCityMembershipRejectsInvalidParameters(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "missing year", path: "/relcensus/city-membership"},
		{name: "non-numeric year", path: "/relcensus/city-membership?year=unknown"},
		{name: "unsupported year", path: "/relcensus/city-membership?year=1925"},
		{
			name: "denomination and family",
			path: "/relcensus/city-membership?year=1926&denomination=Baptist&denominationFamily=Baptist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()

			New(nil).RelCensusCityMembershipHandler().ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
		})
	}
}
