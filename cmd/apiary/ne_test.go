package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestGlobe(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ne/globe", nil)
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

	if len(data.Features) != 255 {
		t.Error("Incorrect number of features returned. Got: ", len(data.Features), " Expected: 254")
	}
}

func TestNorthAmerica(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ne/globe?location=North+America", nil)
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

	if len(data.Features) != 38 {
		t.Error("Incorrect number of features returned. Got: ", len(data.Features), " Expected: 38")
	}
}

func TestAsia(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ne/globe?location=Asia", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data GeoJSONFeatureCollection
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if len(data.Features) == 0 {
		t.Error("No features returned.")
	}

	if data.Type != "FeatureCollection" {
		t.Error("Data is not a FeatureCollection.")
	}

	if len(data.Features) != 53 {
		t.Error("Incorrect number of features returned. Got: ", len(data.Features), " Expected: 53")
	}
}
