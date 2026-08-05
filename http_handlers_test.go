package apiary

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestEndpointsHandler(t *testing.T) {
	tests := []struct {
		name       string
		requestURL string
		tls        bool
		wantPrefix string
	}{
		{
			name:       "http",
			requestURL: "http://data.example/",
			wantPrefix: "http://data.example/",
		},
		{
			name:       "https",
			requestURL: "https://data.example/",
			tls:        true,
			wantPrefix: "https://data.example/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.requestURL, nil)
			if tt.tls {
				request.TLS = &tls.ConnectionState{}
			}
			response := httptest.NewRecorder()

			(&Server{}).EndpointsHandler().ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
			var endpoints []Endpoint
			if err := json.Unmarshal(response.Body.Bytes(), &endpoints); err != nil {
				t.Fatalf("unmarshal endpoint index: %v", err)
			}
			if len(endpoints) != 41 {
				t.Fatalf("endpoint count = %d, want 41", len(endpoints))
			}
			for _, endpoint := range endpoints {
				if !strings.HasPrefix(endpoint.URL, tt.wantPrefix) {
					t.Fatalf(
						"endpoint URL %q does not start with %q",
						endpoint.URL,
						tt.wantPrefix,
					)
				}
			}
			if strings.Contains(response.Body.String(), `\u0026`) {
				t.Fatal("endpoint index escaped query-string ampersands")
			}
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if origin := response.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want *", origin)
	}
}

func TestRoutesRegistersHandlersAndNotFound(t *testing.T) {
	server := &Server{Router: mux.NewRouter()}
	server.Routes()

	for _, path := range []string{
		"/ahcb/counties/1844-05-08/",
		"/apb/bible-books",
		"/bom/parishes",
		"/pinkertons/activities/1",
		"/relcensus/cities",
	} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		if matched := server.Router.Match(request, &mux.RouteMatch{}); !matched {
			t.Errorf("GET route %q was not registered", path)
		}
	}

	post := httptest.NewRequest(http.MethodPost, "/apb/bible-books", nil)
	if matched := server.Router.Match(post, &mux.RouteMatch{}); matched {
		t.Fatal("POST unexpectedly matched a GET/HEAD-only route")
	}

	response := httptest.NewRecorder()
	server.Router.ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/not-a-route", nil),
	)
	if response.Code != http.StatusNotFound {
		t.Fatalf("not-found status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if body := strings.TrimSpace(response.Body.String()); body != "404 Not found." {
		t.Fatalf("not-found body = %q, want %q", body, "404 Not found.")
	}
}
