package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/religious-ecologies/relecapi"
)

var s *relecapi.Server

// Basic structure of a FeatureCollection in GeoJSON
type GeoJSONFeatureCollection struct {
	Type     string        `json:"type"`
	Features []interface{} `json:"features"`
}

func TestMain(m *testing.M) {
	os.Setenv("RELECAPI_LOGGING", "off") // No logs during testing
	s = relecapi.NewServer()
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
func TestSources(t *testing.T) {
	// It is sufficient to check that the list of endpoints is there
	req, _ := http.NewRequest("GET", "/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestPresbyterians(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/presbyterians/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []relecapi.PresbyteriansByYear
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	// Check that the data has the right content
	expected := []relecapi.PresbyteriansByYear{
		{Year: 1826, Members: 127440, Churches: 1819},
		{Year: 1827, Members: 135285, Churches: 1887}}
	if !reflect.DeepEqual(data[0:2], expected) {
		t.Error("Values in data are not what was expected.")
	}

}

func TestCatholicDioceses(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/catholic-dioceses/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []relecapi.CatholicDiocese
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	// Check that the data has the right content
	expected := []relecapi.CatholicDiocese{
		{City: "San Juan", State: "Puerto Rico", Country: "USA",
			Rite: "Latin", DateErected: "1511-08-11", DateMetropolitan: "1960-04-30",
			Lon: -66.1057, Lat: 18.4663},
		{City: "Puebla de los Ángeles", State: "Puebla", Country: "MEX",
			Rite: "Latin", DateErected: "1525-10-13", DateMetropolitan: "1903-08-11",
			Lon: -98.2062, Lat: 19.0413}}
	if !reflect.DeepEqual(data[0:2], expected) {
		t.Error("Values in data are not what was expected.")
	}

}
func TestAHCBStates(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ahcb/states/1789-07-04/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	type GeoJSONFeatureCollection struct {
		Type     string        `json:"type"`
		Features []interface{} `json:"features"`
	}
	var data GeoJSONFeatureCollection
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if data.Type != "FeatureCollection" {
		t.Error("Data is not a FeatureCollection.")
	}

	if len(data.Features) != 16 {
		t.Error("Incorrect number of counties returned.")
	}

}

func TestAHCBCounties(t *testing.T) {
	req, _ := http.NewRequest("GET", "/ahcb/counties/1926-07-04/", nil)
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

	if len(data.Features) != 3113 {
		t.Error("Incorrect number of features returned.")
	}
}

func TestAHCBCountiesByID(t *testing.T) {
	req, _ := http.NewRequest("GET",
		"/ahcb/counties/1980-12-31/id/vas_fairfax,vas_arlington,vas_princewilliam/", nil)
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

	if len(data.Features) != 3 {
		t.Error("Incorrect number of features returned.")
	}
}

func TestAHCBCountiesByStateTerrId(t *testing.T) {
	req, _ := http.NewRequest("GET",
		"/ahcb/counties/1980-12-31/state-terr-id/ga_state,va_state/", nil)
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

	if len(data.Features) != 295 {
		t.Error("Incorrect number of features returned.")
	}
}

func TestAHCBCountiesByStateCode(t *testing.T) {
	req, _ := http.NewRequest("GET",
		"/ahcb/counties/1940-12-31/state-code/nd,sd/", nil)
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

	if len(data.Features) != 122 {
		t.Error("Incorrect number of features returned.")
	}
}

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

	if len(data.Features) != 17 {
		t.Error("Incorrect number of features returned.")
	}
}

func Test404(t *testing.T) {
	req, _ := http.NewRequest("GET", "/nodatahere/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}
