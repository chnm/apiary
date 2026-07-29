package main

import (
	"net/http"
	"net/url"
	"testing"

	apiary "github.com/chnm/apiary"
)

func TestRelCensusDenominations(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/relcensus/denominations", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[[]apiary.Denomination](t, response)
	if len(data) == 0 {
		t.Fatal("expected Religious Census denominations")
	}
}

func TestRelCensusFamilies(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/relcensus/denomination-families", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[struct {
		FamilyRelec []apiary.DenominationFamily `json:"family_relec"`
	}](t, response)
	if len(data.FamilyRelec) == 0 {
		t.Fatal("expected Religious Census denomination families")
	}

	for index, family := range data.FamilyRelec {
		if family.Name == "" {
			t.Errorf("family %d has an empty name", index)
		}
		if index > 0 && family.Name < data.FamilyRelec[index-1].Name {
			t.Errorf(
				"families are not sorted at rows %d and %d: %q, %q",
				index-1,
				index,
				data.FamilyRelec[index-1].Name,
				family.Name,
			)
		}
	}
}

func fetchRelCensusDenominations(t *testing.T) []apiary.Denomination {
	t.Helper()

	request, _ := http.NewRequest(http.MethodGet, "/relcensus/denominations", nil)
	response := executeRequest(request)
	checkResponseCode(t, http.StatusOK, response.Code)
	data := decodeResponse[[]apiary.Denomination](t, response)
	if len(data) == 0 {
		t.Fatal("expected Religious Census denominations")
	}
	return data
}

func fetchRelCensusFamilies(t *testing.T) []apiary.DenominationFamily {
	t.Helper()

	request, _ := http.NewRequest(http.MethodGet, "/relcensus/denomination-families", nil)
	response := executeRequest(request)
	checkResponseCode(t, http.StatusOK, response.Code)
	data := decodeResponse[struct {
		FamilyRelec []apiary.DenominationFamily `json:"family_relec"`
	}](t, response)
	if len(data.FamilyRelec) == 0 {
		t.Fatal("expected Religious Census denomination families")
	}
	return data.FamilyRelec
}

func TestRelCensusCityDenominations(t *testing.T) {
	// Check that we get the right response
	denomination := fetchRelCensusDenominations(t)[0].Name
	req, _ := http.NewRequest(
		http.MethodGet,
		"/relcensus/city-membership?year=1926&denomination="+url.QueryEscape(denomination),
		nil,
	)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[[]apiary.CityMembership](t, response)
	for index, row := range data {
		if row.Year != 1926 || row.Group != denomination {
			t.Errorf(
				"row %d = year %d group %q, want 1926 and %q",
				index,
				row.Year,
				row.Group,
				denomination,
			)
		}
	}
}

func TestRelCensusCityFamilies(t *testing.T) {
	// Check that we get the right response
	family := fetchRelCensusFamilies(t)[0].Name
	req, _ := http.NewRequest(
		http.MethodGet,
		"/relcensus/city-membership?year=1926&denominationFamily="+url.QueryEscape(family),
		nil,
	)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[[]apiary.CityMembership](t, response)
	for index, row := range data {
		if row.Year != 1926 || row.Group != family {
			t.Errorf(
				"row %d = year %d group %q, want 1926 and %q",
				index,
				row.Year,
				row.Group,
				family,
			)
		}
	}
}

func TestRelCensusCityAggregates(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/relcensus/city-membership?year=1926", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[[]apiary.CityMembership](t, response)
	if len(data) == 0 {
		t.Fatal("expected aggregated Religious Census city membership")
	}
	for index, row := range data {
		if row.Year != 1926 || row.Group != "All denominations" {
			t.Errorf(
				"row %d = year %d group %q, want 1926 and All denominations",
				index,
				row.Year,
				row.Group,
			)
		}
	}
}

func TestRelCensusLocations(t *testing.T) {
	req, _ := http.NewRequest("GET", "/relcensus/cities", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[[]apiary.LocationInfo](t, response)
	if len(data) == 0 {
		t.Fatal("expected Religious Census locations")
	}
}
