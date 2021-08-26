package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/chnm/dataapi"
)

var s *dataapi.Server

// Basic structure of a FeatureCollection in GeoJSON
type GeoJSONFeatureCollection struct {
	Type     string        `json:"type"`
	Features []interface{} `json:"features"`
}

func TestMain(m *testing.M) {
	os.Setenv("DATAAPI_LOGGING", "off") // No logs during testing
	s = dataapi.NewServer()
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

func TestPresbyterians(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/presbyterians/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []dataapi.PresbyteriansByYear
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	// Check that the data has the right content
	expected := []dataapi.PresbyteriansByYear{
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
	var data []dataapi.CatholicDiocese
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}

func TestCatholicDiocesesPerDecade(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/catholic-dioceses/per-decade/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []dataapi.CatholicDiocesesPerDecade
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	// Check that the data has the right content
	expected := []dataapi.CatholicDiocesesPerDecade{
		{Decade: 1500, Count: 0},
		{Decade: 1510, Count: 3},
		{Decade: 1520, Count: 1},
	}
	if !reflect.DeepEqual(data[0:3], expected) {
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

	if len(data.Features) != 37 {
		t.Error("Incorrect number of features returned.")
	}
}

func TestCountiesInState(t *testing.T) {
	req, _ := http.NewRequest("GET", "/pop-places/state/nc/county/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []dataapi.PlaceCounty
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	// Check that the data has the right content
	expected := []dataapi.PlaceCounty{
		{CountyAHCB: "ncs_alamance", County: "Alamance"},
		{CountyAHCB: "ncs_alexander", County: "Alexander"},
	}
	if !reflect.DeepEqual(data[0:2], expected) {
		t.Error("Values in data are not what was expected.")
	}

}

func TestPlacesInCounty(t *testing.T) {
	req, _ := http.NewRequest("GET", "/pop-places/county/mas_middlesex/place/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []dataapi.Place
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

}

func TestPlace(t *testing.T) {
	req, _ := http.NewRequest("GET", "/pop-places/place/611119/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data dataapi.PlaceDetails
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	expected := dataapi.PlaceDetails{
		PlaceID:    611119,
		Place:      "Groton",
		MapName:    "Ayer",
		County:     "Middlesex",
		CountyAHCB: "mas_middlesex",
		State:      "MA",
	}
	if !reflect.DeepEqual(data, expected) {
		t.Error("Values in data are not what was expected.")
	}

}

func TestAPBFeaturedVerses(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/index/featured", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []dataapi.APBIndexItem
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if len(data) < 4 {
		t.Error("Not enough verses returned.")
	}

}

func TestAPBTopVerses(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/index/top", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []dataapi.APBIndexItem
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if len(data) < 100 {
		t.Error("Not enough verses returned.")
	}

}

func TestAPBVersePeaks(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/index/peaks", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []dataapi.APBIndexItem
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if len(data) < 100 {
		t.Error("Not enough verses returned.")
	}

}

func TestAPBVerse(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/verse?ref=Genesis+1:1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data dataapi.Verse
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if data.Reference != "Genesis 1:1" {
		t.Error("Wrong verse returned.")
	}

}

func TestAPBVerseTrend(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/verse-trend?ref=Genesis+1:1&corpus=chronam", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data dataapi.VerseTrendResponse
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if data.Reference != "Genesis 1:1" {
		t.Error("Wrong verse returned.")
	}

	if data.Corpus != "chronam" {
		t.Error("Wrong corpus returned.")
	}

	if len(data.Trend) < 50 {
		t.Error("Not enough data points returned.")
	}

}

func TestAPBBibleTrend(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/bible-trend", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data dataapi.VerseTrendResponse
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if data.Reference != "bible" {
		t.Error("Wrong verse returned.")
	}

	if data.Corpus != "chronam" {
		t.Error("Wrong corpus returned.")
	}

	if len(data.Trend) < 50 {
		t.Error("Not enough data points returned.")
	}

}

func TestAPBVerseQuotations(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/verse-quotations?ref=Genesis+1:1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []dataapi.VerseQuotation
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if len(data) < 100 {
		t.Error("Not enough verses returned.")
	}

}
