package main

import (
	"net/http"
	"testing"
)

func TestAHCBStates(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ahcb/states/1789-07-04/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[GeoJSONFeatureCollection](t, response)
	requireNonEmptyFeatureCollection(t, data)
}

func TestAHCBCounties(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ahcb/counties/1926-07-04/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[GeoJSONFeatureCollection](t, response)
	requireNonEmptyFeatureCollection(t, data)
}

func TestAHCBCountiesByID(t *testing.T) {
	req, _ := http.NewRequest("GET",
		"/ahcb/counties/1980-12-31/id/vas_fairfax,vas_arlington,vas_princewilliam/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[GeoJSONFeatureCollection](t, response)
	requireNonEmptyFeatureCollection(t, data)
}

func TestAHCBCountiesByStateTerrId(t *testing.T) {
	req, _ := http.NewRequest("GET",
		"/ahcb/counties/1980-12-31/state-terr-id/ga_state,va_state/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[GeoJSONFeatureCollection](t, response)
	requireNonEmptyFeatureCollection(t, data)
}

func TestAHCBCountiesByStateCode(t *testing.T) {
	req, _ := http.NewRequest("GET",
		"/ahcb/counties/1940-12-31/state-code/nd,sd/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[GeoJSONFeatureCollection](t, response)
	requireNonEmptyFeatureCollection(t, data)
}
