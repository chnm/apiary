package apb

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVerseTrendRejectsInvalidParameters(t *testing.T) {
	tests := []struct {
		name string
		path string
		body string
	}{
		{
			name: "invalid corpus",
			path: "/apb/verse-trend?ref=Gen.1.1&corpus=unknown",
			body: "400 Bad request. Corpus must be 'ncnp' or 'chronam'.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()

			New(nil).APBVerseTrendHandler().ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
			if body := strings.TrimSpace(response.Body.String()); body != tt.body {
				t.Fatalf("body = %q, want %q", body, tt.body)
			}
		})
	}
}
