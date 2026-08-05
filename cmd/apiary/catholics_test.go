package main

import (
	"net/http"
	"testing"

	"github.com/chnm/apiary/internal/datasets/catholic"
)

func TestCatholicDioceses(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/catholic-dioceses/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[[]catholic.CatholicDiocese](t, response)
	if len(data) == 0 {
		t.Fatal("expected Catholic dioceses")
	}
	for index, diocese := range data {
		if diocese.City == "" || diocese.Country == "" || diocese.Rite == "" {
			t.Errorf("diocese %d is missing identifying fields: %+v", index, diocese)
		}
	}
}

func TestCatholicDiocesesPerDecade(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/catholic-dioceses/per-decade/", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	data := decodeResponse[[]catholic.CatholicDiocesesPerDecade](t, response)
	if len(data) == 0 {
		t.Fatal("expected Catholic diocese decade statistics")
	}

	for index, row := range data {
		if row.Decade%10 != 0 {
			t.Errorf("row %d has non-decade year %d", index, row.Decade)
		}
		if row.Count < 0 {
			t.Errorf("row %d has negative count %d", index, row.Count)
		}
		if index > 0 && row.Decade != data[index-1].Decade+10 {
			t.Errorf(
				"decades are not consecutive at rows %d and %d: %d, %d",
				index-1,
				index,
				data[index-1].Decade,
				row.Decade,
			)
		}
	}
}
