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

func TestBomBills(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/bom/bills?start-year=1669&end-year=1754&bill-type=All&count-type=All&limit=50&offset=0", nil)
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
	req, _ := http.NewRequest("GET", "/bom/christenings?start-year=1669&end-year=1754", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.ChristeningsByYear
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}

func TestBomCauses(t *testing.T) {
	req, _ := http.NewRequest("GET", "/bom/causes?start-year=1669&end-year=1754", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var data []apiary.DeathCauses
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}

func TestBomListChristenings(t *testing.T) {
	req, _ := http.NewRequest("GET", "/bom/list-christenings", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var data []apiary.Christenings
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}

func TestBomListCauses(t *testing.T) {
	req, _ := http.NewRequest("GET", "/bom/list-deaths", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var data []apiary.Causes
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}
