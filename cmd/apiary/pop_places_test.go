package main

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	apiary "github.com/chnm/apiary"
)

func TestCountiesInState(t *testing.T) {
	req, _ := http.NewRequest("GET", "/pop-places/state/nc/county/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.PlaceCounty
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	// Check that the data has the right content
	expected := []apiary.PlaceCounty{
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
	var data []apiary.Place
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
	var data apiary.PlaceDetails
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	expected := apiary.PlaceDetails{
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
