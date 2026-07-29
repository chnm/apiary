package main

import (
	"net/http"
	"testing"
)

func TestGlobe(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ne/globe", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[GeoJSONFeatureCollection](t, response)
	requireNonEmptyFeatureCollection(t, data)
}

func TestNorthAmerica(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ne/globe?location=North+America", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[GeoJSONFeatureCollection](t, response)
	requireNonEmptyFeatureCollection(t, data)
}

func TestAsia(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ne/globe?location=Asia", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[GeoJSONFeatureCollection](t, response)
	requireNonEmptyFeatureCollection(t, data)
}
