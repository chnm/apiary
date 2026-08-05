package apiary

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

func TestClientCacheMiddleware(t *testing.T) {
	// Successful responses are client-cacheable for a week.
	ok := clientCacheMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	rr := httptest.NewRecorder()
	ok.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if got := rr.Header().Get("Cache-Control"); got != "max-age=604800" {
		t.Errorf("success Cache-Control = %q, want max-age=604800", got)
	}

	// Error responses must not be cached by clients.
	bad := clientCacheMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	rr = httptest.NewRecorder()
	bad.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if got := rr.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("error Cache-Control = %q, want no-store", got)
	}
}

func TestResponseCacheMiddleware(t *testing.T) {
	adapter, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(1024),
	)
	if err != nil {
		t.Fatalf("create cache adapter: %v", err)
	}
	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(adapter),
		cache.ClientWithTTL(time.Hour),
		cache.ClientWithRefreshKey("nocache"),
	)
	if err != nil {
		t.Fatalf("create cache client: %v", err)
	}

	requests := 0
	router := mux.NewRouter()
	router.HandleFunc("/data", func(w http.ResponseWriter, _ *http.Request) {
		requests++
		_, _ = w.Write([]byte(strconv.Itoa(requests)))
	})
	server := &Server{Router: router, Cache: cacheClient}
	server.Middleware()

	execute := func(path string) string {
		t.Helper()
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", path, response.Code, http.StatusOK)
		}
		return response.Body.String()
	}

	if body := execute("/data"); body != "1" {
		t.Fatalf("first response = %q, want 1", body)
	}
	if body := execute("/data"); body != "1" {
		t.Fatalf("cached response = %q, want 1", body)
	}
	if requests != 1 {
		t.Fatalf("handler calls after cached request = %d, want 1", requests)
	}
	if body := execute("/data?nocache"); body != "2" {
		t.Fatalf("refreshed response = %q, want 2", body)
	}
	if requests != 2 {
		t.Fatalf("handler calls after refresh = %d, want 2", requests)
	}
}
