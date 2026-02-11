package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestEndpoints(t *testing.T) {
	// It is sufficient to check that the list of endpoints is there
	req, _ := http.NewRequest("GET", "/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
}

func Test404(t *testing.T) {
	req, _ := http.NewRequest("GET", "/nodatahere/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestCache(t *testing.T) {
	req, _ := http.NewRequest("GET", "/cache", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data and verify structure
	var data struct {
		Startup string `json:"startup"`
		Handler string `json:"handler"`
	}
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	// Verify the response has the expected fields
	if data.Startup == "" {
		t.Error("Expected startup time to be set")
	}
	if data.Handler == "" {
		t.Error("Expected handler time to be set")
	}
}
