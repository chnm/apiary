package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
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

	if len(data) == 0 {
		t.Fatal("expected BOM parishes")
	}

	seenIDs := make(map[int]bool, len(data))
	for index, parish := range data {
		if parish.ParishID <= 0 || parish.Name == "" || parish.CanonicalName == "" {
			t.Errorf("parish %d is missing identifying fields: %+v", index, parish)
		}
		if seenIDs[parish.ParishID] {
			t.Errorf("parish ID %d appears more than once", parish.ParishID)
		}
		seenIDs[parish.ParishID] = true
	}
}

func TestBomBills(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/bom/bills?start-year=1669&end-year=1754&bill-type=Weekly&count-type=Buried&limit=50&offset=0", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data apiary.PaginatedResponse
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
	if data.Data == nil {
		t.Error("expected paginated bill data, got nil")
	}
}

func TestBomTotalBills(t *testing.T) {
	req, _ := http.NewRequest("GET", "/bom/totalbills?type=weekly", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var data []apiary.TotalBills
	if err := json.Unmarshal(response.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected total bill data")
	}
}

func TestBomStatistics(t *testing.T) {
	req, _ := http.NewRequest("GET", "/bom/statistics?type=yearly", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var data []apiary.YearlySummary
	if err := json.Unmarshal(response.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected yearly statistics")
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

func TestBomShapefiles(t *testing.T) {
	// Test base case with no parameters
	t.Run("BaseRequest", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		var data GeoJSONFeatureCollection
		err := json.Unmarshal(response.Body.Bytes(), &data)
		if err != nil {
			t.Error(err)
		}

		// Verify it's a GeoJSON FeatureCollection
		if data.Type != "FeatureCollection" {
			t.Errorf("Expected type FeatureCollection, got %s", data.Type)
		}
	})

	// Test with year parameter
	t.Run("YearFilter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?year=1665", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		var data GeoJSONFeatureCollection
		err := json.Unmarshal(response.Body.Bytes(), &data)
		if err != nil {
			t.Error(err)
		}
	})

	// Test with date range
	t.Run("DateRangeFilter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?start-year=1660&end-year=1670", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		var data GeoJSONFeatureCollection
		err := json.Unmarshal(response.Body.Bytes(), &data)
		if err != nil {
			t.Error(err)
		}
	})

	// Test with bill type
	t.Run("BillTypeFilter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?bill-type=Weekly", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		var data GeoJSONFeatureCollection
		err := json.Unmarshal(response.Body.Bytes(), &data)
		if err != nil {
			t.Error(err)
		}
	})

	// Test with count type
	t.Run("CountTypeFilter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?count-type=Buried", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		var data GeoJSONFeatureCollection
		err := json.Unmarshal(response.Body.Bytes(), &data)
		if err != nil {
			t.Error(err)
		}
	})

	// Test with subunit
	t.Run("SubunitFilter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?subunit=City", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		var data GeoJSONFeatureCollection
		err := json.Unmarshal(response.Body.Bytes(), &data)
		if err != nil {
			t.Error(err)
		}
	})

	// Test with city_cnty
	t.Run("CityCntyFilter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?city_cnty=London", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		var data GeoJSONFeatureCollection
		err := json.Unmarshal(response.Body.Bytes(), &data)
		if err != nil {
			t.Error(err)
		}
	})

	// Test with parish
	t.Run("ParishFilter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?parish=1,2", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		var data GeoJSONFeatureCollection
		err := json.Unmarshal(response.Body.Bytes(), &data)
		if err != nil {
			t.Error(err)
		}
	})

	// Test with multiple filters
	t.Run("MultipleFilters", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?year=1665&count-type=Plague&city_cnty=London", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		var data GeoJSONFeatureCollection
		err := json.Unmarshal(response.Body.Bytes(), &data)
		if err != nil {
			t.Error(err)
		}
	})

	// Test invalid bill type
	t.Run("InvalidBillType", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?bill-type=Invalid", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusBadRequest, response.Code)
	})

	// Test invalid count type
	t.Run("InvalidCountType", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?count-type=Invalid", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusBadRequest, response.Code)
	})

	// Test invalid parish ID format
	t.Run("InvalidParishID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?parish=abc", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusBadRequest, response.Code)
	})

	// Test with invalid year format
	t.Run("InvalidYearFormat", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?year=abc", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusBadRequest, response.Code)
	})

	// Test content type
	t.Run("ContentType", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles", nil)
		response := executeRequest(req)
		contentType := response.Header().Get("Content-Type")
		expectedContentType := "application/geo+json"

		if contentType != expectedContentType {
			t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
		}
	})

	// Test cache headers
	t.Run("CacheHeaders", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles", nil)
		response := executeRequest(req)

		cacheControl := response.Header().Get("Cache-Control")
		if cacheControl != "public, max-age=86400" {
			t.Errorf("Expected Cache-Control: public, max-age=86400, got %s", cacheControl)
		}

		vary := response.Header().Get("Vary")
		if vary != "Accept-Encoding" {
			t.Errorf("Expected Vary: Accept-Encoding, got %s", vary)
		}
	})
}

// Helper tests for the validation functions
func TestIsValidBillType(t *testing.T) {
	// Test with valid bill types (case-insensitive)
	t.Run("ValidBillTypes", func(t *testing.T) {
		validTypes := []string{"Weekly", "General", "Total", "weekly", "WEEKLY", "general", "GENERAL", "total", "TOTAL"}
		for _, billType := range validTypes {
			if !apiary.IsValidBillType(billType) {
				t.Errorf("Expected %s to be a valid bill type, but it was rejected", billType)
			}
		}
	})

	// Test with invalid bill types
	t.Run("InvalidBillTypes", func(t *testing.T) {
		invalidTypes := []string{"", "Invalid", "MonthlyReport", "monthly", "Daily"}
		for _, billType := range invalidTypes {
			if apiary.IsValidBillType(billType) {
				t.Errorf("Expected %s to be an invalid bill type, but it was accepted", billType)
			}
		}
	})
}

func TestIsValidCountType(t *testing.T) {
	// Test with valid count types (case-insensitive)
	t.Run("ValidCountTypes", func(t *testing.T) {
		validTypes := []string{"Buried", "Plague", "buried", "BURIED", "plague", "PLAGUE"}
		for _, countType := range validTypes {
			if !apiary.IsValidCountType(countType) {
				t.Errorf("Expected %s to be a valid count type, but it was rejected", countType)
			}
		}
	})

	// Test with invalid count types
	t.Run("InvalidCountTypes", func(t *testing.T) {
		invalidTypes := []string{"", "Invalid", "Deaths", "Christenings", "dead", "sick"}
		for _, countType := range invalidTypes {
			if apiary.IsValidCountType(countType) {
				t.Errorf("Expected %s to be an invalid count type, but it was accepted", countType)
			}
		}
	})
}

func TestBomBillsInvalidParish(t *testing.T) {
	// Unknown parish IDs return 400 naming the offending IDs.
	req, _ := http.NewRequest(
		"GET",
		"/bom/bills?start-year=1669&end-year=1754&parish=2147483647",
		nil,
	)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
	if body := strings.TrimSpace(response.Body.String()); body != "invalid parish ID(s): 2147483647" {
		t.Errorf("unexpected error body: %q", body)
	}

	// Multiple unknown IDs are all reported, comma-space separated, in request order.
	req, _ = http.NewRequest(
		"GET",
		"/bom/bills?start-year=1669&end-year=1754&parish=2147483646,2147483647",
		nil,
	)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
	if body := strings.TrimSpace(response.Body.String()); body != "invalid parish ID(s): 2147483646, 2147483647" {
		t.Errorf("unexpected error body: %q", body)
	}

	// Select a valid parish ID from the database rather than assuming one.
	parishRequest, _ := http.NewRequest(http.MethodGet, "/bom/parishes", nil)
	parishResponse := executeRequest(parishRequest)
	checkResponseCode(t, http.StatusOK, parishResponse.Code)
	parishes := decodeResponse[[]apiary.Parish](t, parishResponse)
	if len(parishes) == 0 {
		t.Fatal("expected at least one parish")
	}

	req, _ = http.NewRequest(
		"GET",
		"/bom/bills?start-year=1669&end-year=1754&parish="+
			strconv.Itoa(parishes[0].ParishID),
		nil,
	)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
}
