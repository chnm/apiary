package apiary

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestTotalBillsHandlerRejectsInvalidType(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/bom/totalbills?type=unknown", nil)
	response := httptest.NewRecorder()

	(&Server{}).TotalBillsHandler().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if body := strings.TrimSpace(response.Body.String()); !strings.HasPrefix(body, "Invalid type parameter.") {
		t.Fatalf("body = %q, want invalid type error", body)
	}
}

func TestParseAPIParametersParish(t *testing.T) {
	// Non-integer parish IDs are rejected by the parser.
	r := httptest.NewRequest("GET", "/bom/bills?parish=abc", nil)
	if _, err := parseAPIParameters(r); err == nil {
		t.Error("expected error for non-integer parish ID, got nil")
	}

	// Comma-separated parish IDs are parsed in order, with whitespace trimmed.
	r = httptest.NewRequest("GET", "/bom/bills?parish=1,%205,151", nil)
	params, err := parseAPIParameters(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(params.Parish, []int{1, 5, 151}) {
		t.Errorf("expected [1 5 151], got %v", params.Parish)
	}

	// The parser does not range-check IDs; existence is validated against
	// the database by Server.invalidParishIDs.
	r = httptest.NewRequest("GET", "/bom/bills?parish=0,99999", nil)
	params, err = parseAPIParameters(r)
	if err != nil {
		t.Fatalf("unexpected error for out-of-range IDs: %v", err)
	}
	if !reflect.DeepEqual(params.Parish, []int{0, 99999}) {
		t.Errorf("expected [0 99999], got %v", params.Parish)
	}
}
