package main

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	apiary "github.com/chnm/apiary"
)

func TestBomParishes(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/bom/parishes", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.Parish
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	// Check that the data has the right content
	expected := []apiary.Parish{
		{ParishID: 1, Name: "Alhallows Barking", CanonicalName: "All Hallows Barking"},
		{ParishID: 2, Name: "Alhallows Breadstreet", CanonicalName: "All Hallows Bread Street"}}
	if !reflect.DeepEqual(data[0:2], expected) {
		t.Error("Values in data are not what was expected.")
	}
}

func TestWeeklyBomBills(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/bom/bills?startYear=1669&endYear=1754", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.ParishByYear
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}

func TestGeneralBomBills(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/bom/generalbills?startYear=1669&endYear=1754", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.ParishByYear
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}

func TestBomChristenings(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/bom/christenings?startYear=1669&endYear=1754", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.ParishByYear
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}
