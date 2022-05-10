package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/chnm/apiary"
)

var s *apiary.Server

// Basic structure of a FeatureCollection in GeoJSON
type GeoJSONFeatureCollection struct {
	Type     string        `json:"type"`
	Features []interface{} `json:"features"`
}

func TestMain(m *testing.M) {
	os.Setenv("apiary_LOGGING", "off") // No logs during testing
	s = apiary.NewServer()
	code := m.Run()
	os.Exit(code)
}

// Helper for tests.
func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)
	return rr
}

// Helper for tests.
func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d.\n", expected, actual)
	}
}
