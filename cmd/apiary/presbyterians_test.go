package main

import (
	"net/http"
	"testing"

	apiary "github.com/chnm/apiary"
)

func TestPresbyterians(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/presbyterians/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[[]apiary.PresbyteriansByYear](t, response)
	if len(data) == 0 {
		t.Fatal("expected Presbyterian statistics")
	}

	for index, row := range data {
		if row.Year <= 0 {
			t.Errorf("row %d has invalid year %d", index, row.Year)
		}
		if row.Members < 0 || row.Churches < 0 {
			t.Errorf(
				"row %d has negative totals: members=%d churches=%d",
				index,
				row.Members,
				row.Churches,
			)
		}
		if index > 0 && row.Year <= data[index-1].Year {
			t.Errorf(
				"years are not strictly increasing at rows %d and %d: %d, %d",
				index-1,
				index,
				data[index-1].Year,
				row.Year,
			)
		}
	}
}
