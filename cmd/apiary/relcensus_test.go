package main

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	apiary "github.com/chnm/apiary"
)

func TestRelCensusDenominations(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/relcensus/denominations", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.Denomination
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}

type FamilyMap struct {
	Param map[string]string `json:"family_relec"`
}

func TestRelCensusFamilies(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/relcensus/denomination-families", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	// var data []apiary.DenominationFamily
	data := struct {
		FamilyRelec []apiary.DenominationFamily `json:"family_relec"`
	}{}
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	// Check that the data has the right content
	expected := []apiary.DenominationFamily{
		{Name: "Adventist"},
		{Name: "Anabaptist"},
		{Name: "Baptist"},
	}

	if !reflect.DeepEqual(data.FamilyRelec[0:3], expected) {
		t.Errorf("Expected %v, got %v", expected, data.FamilyRelec)
	}
}

func TestRelCensusCityDenominations(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/relcensus/city-membership?year=1926&denomination=Church+of+God+in+Christ", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.CityMembership
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}

func TestRelCensusCityFamilies(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/relcensus/city-membership?year=1926&denominationFamily=Pentecostal", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.CityMembership
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}

func TestRelCensusCityAggregates(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/relcensus/city-membership?year=1926", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.CityMembership
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}

func TestRelCensusLocations(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/relcensus/cities", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.LocationInfo
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	// Check that we got an array
	if data == nil {
		t.Error("Expected array of locations, got nil")
	}
}
