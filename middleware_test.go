package apiary

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
