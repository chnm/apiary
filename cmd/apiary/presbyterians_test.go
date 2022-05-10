package main

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	apiary "github.com/chnm/apiary"
)

func TestPresbyterians(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/presbyterians/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.PresbyteriansByYear
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	// Check that the data has the right content
	expected := []apiary.PresbyteriansByYear{
		{Year: 1826, Members: 127440, Churches: 1819},
		{Year: 1827, Members: 135285, Churches: 1887}}
	if !reflect.DeepEqual(data[0:2], expected) {
		t.Error("Values in data are not what was expected.")
	}

}
