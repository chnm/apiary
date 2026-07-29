package main

import (
	"context"
	"encoding/json"
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

// The integration suite runs against the database configured by APIARY_DB.
// Assertions should validate response contracts and stable invariants rather
// than exact counts or records from mutable data.
func TestMain(m *testing.M) {
	os.Setenv("APIARY_LOGGING", "off") // No logs during testing
	s = apiary.NewServer(context.TODO())
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
	t.Helper()
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d.\n", expected, actual)
	}
}

func decodeResponse[T any](t *testing.T, response *httptest.ResponseRecorder) T {
	t.Helper()

	var data T
	if err := json.Unmarshal(response.Body.Bytes(), &data); err != nil {
		t.Fatalf("decode response JSON: %v", err)
	}
	return data
}

func requireNonEmptyFeatureCollection(
	t *testing.T,
	data GeoJSONFeatureCollection,
) {
	t.Helper()

	if data.Type != "FeatureCollection" {
		t.Fatalf("GeoJSON type = %q, want FeatureCollection", data.Type)
	}
	if len(data.Features) == 0 {
		t.Fatal("GeoJSON FeatureCollection is empty")
	}
}
