package main

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/chnm/apiary/internal/datasets/popplaces"
)

func TestCountiesInState(t *testing.T) {
	data := fetchCountiesInState(t, "ma")
	for index, county := range data {
		if county.CountyAHCB == "" || county.County == "" {
			t.Errorf("county %d is missing identifying fields: %+v", index, county)
		}
		if index > 0 && county.County < data[index-1].County {
			t.Errorf(
				"counties are not sorted at rows %d and %d: %q, %q",
				index-1,
				index,
				data[index-1].County,
				county.County,
			)
		}
	}
}

func TestPlacesInCounty(t *testing.T) {
	county := fetchCountiesInState(t, "ma")[0]
	data := fetchPlacesInCounty(t, county.CountyAHCB)
	for index, place := range data {
		if place.PlaceID <= 0 || place.Place == "" || place.Lat == 0 || place.Lon == 0 {
			t.Errorf("place %d is missing identifying fields: %+v", index, place)
		}
	}
}

func TestPlace(t *testing.T) {
	county := fetchCountiesInState(t, "ma")[0]
	place := fetchPlacesInCounty(t, county.CountyAHCB)[0]
	req, _ := http.NewRequest(
		http.MethodGet,
		"/pop-places/place/"+strconv.Itoa(place.PlaceID)+"/",
		nil,
	)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[popplaces.PlaceDetails](t, response)
	if data.PlaceID != place.PlaceID ||
		data.Place != place.Place ||
		data.Lat != place.Lat ||
		data.Lon != place.Lon ||
		data.CountyAHCB != county.CountyAHCB {
		t.Fatalf(
			"place details = %+v, want list record %+v in county %+v",
			data,
			place,
			county,
		)
	}
}

func TestPlaceNotFound(t *testing.T) {
	req, _ := http.NewRequest("GET", "/pop-places/place/2147483647/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func fetchCountiesInState(t *testing.T, state string) []popplaces.PlaceCounty {
	t.Helper()

	request, _ := http.NewRequest(
		http.MethodGet,
		"/pop-places/state/"+state+"/county/",
		nil,
	)
	response := executeRequest(request)
	checkResponseCode(t, http.StatusOK, response.Code)
	data := decodeResponse[[]popplaces.PlaceCounty](t, response)
	if len(data) == 0 {
		t.Fatalf("expected populated-place counties for state %q", state)
	}
	return data
}

func fetchPlacesInCounty(t *testing.T, county string) []popplaces.Place {
	t.Helper()

	request, _ := http.NewRequest(
		http.MethodGet,
		"/pop-places/county/"+county+"/place/",
		nil,
	)
	response := executeRequest(request)
	checkResponseCode(t, http.StatusOK, response.Code)
	data := decodeResponse[[]popplaces.Place](t, response)
	if len(data) == 0 {
		t.Fatalf("expected populated places for county %q", county)
	}
	return data
}
