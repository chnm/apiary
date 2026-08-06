package apb

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVerseHandlersRequireExactlyOneReference(t *testing.T) {
	handler := New(nil)
	handlers := []struct {
		name    string
		path    string
		handler http.HandlerFunc
	}{
		{name: "verse", path: "/apb/verse", handler: handler.APBVerseHandler()},
		{name: "verse quotations", path: "/apb/verse-quotations", handler: handler.APBVerseQuotationsHandler()},
		{name: "verse trend", path: "/apb/verse-trend", handler: handler.APBVerseTrendHandler()},
	}
	queries := []struct {
		name  string
		query string
	}{
		{name: "missing", query: ""},
		{name: "repeated", query: "?ref=Genesis+1%3A1&ref=John+1%3A1"},
	}

	for _, endpoint := range handlers {
		for _, query := range queries {
			t.Run(endpoint.name+"/"+query.name, func(t *testing.T) {
				request := httptest.NewRequest(http.MethodGet, endpoint.path+query.query, nil)
				response := httptest.NewRecorder()

				endpoint.handler.ServeHTTP(response, request)

				if response.Code != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
				}
				if body := strings.TrimSpace(response.Body.String()); body != singleReferenceError {
					t.Fatalf("body = %q, want %q", body, singleReferenceError)
				}
			})
		}
	}
}
