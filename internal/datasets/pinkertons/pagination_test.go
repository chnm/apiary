package pinkertons

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestActivitiesRejectInvalidPaginationParameters(t *testing.T) {
	tests := []struct {
		name string
		path string
		body string
	}{
		{
			name: "zero limit",
			path: "/pinkertons/activities?limit=0",
			body: "Invalid limit parameter",
		},
		{
			name: "non-numeric limit",
			path: "/pinkertons/activities?limit=ten",
			body: "Invalid limit parameter",
		},
		{
			name: "negative offset",
			path: "/pinkertons/activities?offset=-1",
			body: "Invalid offset parameter",
		},
		{
			name: "non-numeric offset",
			path: "/pinkertons/activities?offset=ten",
			body: "Invalid offset parameter",
		},
		{
			name: "zero location ID",
			path: "/pinkertons/activities?location_id=0",
			body: "Invalid location_id parameter",
		},
		{
			name: "non-numeric location ID",
			path: "/pinkertons/activities?location_id=unknown",
			body: "Invalid location_id parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()

			(&Handler{}).ActivitiesHandler().ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf(
					"status = %d, want %d",
					response.Code,
					http.StatusBadRequest,
				)
			}
			if body := strings.TrimSpace(response.Body.String()); body != tt.body {
				t.Fatalf("body = %q, want %q", body, tt.body)
			}
		})
	}
}
