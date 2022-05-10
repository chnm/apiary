package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestNorthAmerica(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ne/northamerica/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data GeoJSONFeatureCollection
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if data.Type != "FeatureCollection" {
		t.Error("Data is not a FeatureCollection.")
	}

	if len(data.Features) != 37 {
		t.Error("Incorrect number of features returned.")
	}
}
