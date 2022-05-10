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

	// Check that the data has the right content
	expected := []apiary.Denomination{
		{
			Name:           "Duck River and Kindred Associations of Baptists (Baptist Church of Christ)",
			ShortName:      "Duck River and Kindred Associations of Baptists",
			DenominationID: "0-2-2",
			FamilyCensus:   "Baptist bodies",
			FamilyRelec:    "Baptist",
		},
		{
			Name:           "Advent Christian Church",
			ShortName:      "Advent Christian Church",
			DenominationID: "0-0-0",
			FamilyCensus:   "Adventist bodies",
			FamilyRelec:    "Adventist",
		},
	}
	if !reflect.DeepEqual(data[0:2], expected) {
		t.Error("Values in data are not what was expected.")
	}
}

// func TestRelCensusFamilies(t *testing.T) {
// 	// Check that we get the right response
// 	req, _ := http.NewRequest("GET", "/relcensus/denomination-families", nil)
// 	response := executeRequest(req)
// 	checkResponseCode(t, http.StatusOK, response.Code)

// 	// Get the data
// 	var data []apiary.DenominationFamily
// 	err := json.Unmarshal(response.Body.Bytes(), &data)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	// Check that the data has the right content
// 	expected := []apiary.DenominationFamily{
// 		{Name: "Adventist"},
// 		{Name: "Anabaptist"},
// 		{Name: "Baptist"},
// 	}
// 	if !reflect.DeepEqual(data[0:3], expected) {
// 		t.Error("Values in data are not what was expected.")
// 	}
// }

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
