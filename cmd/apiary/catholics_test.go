package main

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	apiary "github.com/chnm/apiary"
)

func TestCatholicDioceses(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/catholic-dioceses/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.CatholicDiocese
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
	var data []apiary.CatholicDiocesesPerDecade
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	// Check that the data has the right content
	expected := []apiary.CatholicDiocesesPerDecade{
		{Decade: 1500, Count: 0},
		{Decade: 1510, Count: 3},
		{Decade: 1520, Count: 1},
	}
	if !reflect.DeepEqual(data[0:3], expected) {
		t.Error("Values in data are not what was expected.")
	}
}
