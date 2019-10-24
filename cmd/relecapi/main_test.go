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

func TestMain(m *testing.M) {
	s = relecapi.NewServer()
	code := m.Run()
	os.Exit(code)
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

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d.\n", expected, actual)
	}
}
