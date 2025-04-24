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
	req, _ := http.NewRequest("GET", "/bom/bills?start-year=1669&end-year=1754&bill-type=Weekly&count-type=Buried&limit=50&offset=0", nil)
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

	// Test invalid bill type (should still work but ignore the invalid filter)
	t.Run("InvalidBillType", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?bill-type=Invalid", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		var data GeoJSONFeatureCollection
		err := json.Unmarshal(response.Body.Bytes(), &data)
		if err != nil {
			t.Error(err)
		}
	})

	// Test invalid count type (should still work but ignore the invalid filter)
	t.Run("InvalidCountType", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?count-type=Invalid", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		var data GeoJSONFeatureCollection
		err := json.Unmarshal(response.Body.Bytes(), &data)
		if err != nil {
			t.Error(err)
		}
	})

	// Test invalid parish ID format
	t.Run("InvalidParishID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?parish=abc", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		var data GeoJSONFeatureCollection
		err := json.Unmarshal(response.Body.Bytes(), &data)
		if err != nil {
			t.Error(err)
		}
	})

	// Test with invalid year format
	t.Run("InvalidYearFormat", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/bom/shapefiles?year=abc", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		var data GeoJSONFeatureCollection
		err := json.Unmarshal(response.Body.Bytes(), &data)
		if err != nil {
			t.Error(err)
		}
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
	// Test with valid bill types
	t.Run("ValidBillTypes", func(t *testing.T) {
		validTypes := []string{"Weekly", "General", "Total"}
		for _, billType := range validTypes {
			if !apiary.IsValidBillType(billType) {
				t.Errorf("Expected %s to be a valid bill type, but it was rejected", billType)
			}
		}
	})

	// Test with invalid bill types
	t.Run("InvalidBillTypes", func(t *testing.T) {
		invalidTypes := []string{"", "weekly", "WEEKLY", "Invalid", "MonthlyReport"}
		for _, billType := range invalidTypes {
			if apiary.IsValidBillType(billType) {
				t.Errorf("Expected %s to be an invalid bill type, but it was accepted", billType)
			}
		}
	})
}

func TestIsValidCountType(t *testing.T) {
	// Test with valid count types
	t.Run("ValidCountTypes", func(t *testing.T) {
		validTypes := []string{"Buried", "Plague"}
		for _, countType := range validTypes {
			if !apiary.IsValidCountType(countType) {
				t.Errorf("Expected %s to be a valid count type, but it was rejected", countType)
			}
		}
	})

	// Test with invalid count types
	t.Run("InvalidCountTypes", func(t *testing.T) {
		invalidTypes := []string{"", "buried", "BURIED", "Invalid", "Deaths", "Christenings"}
		for _, countType := range invalidTypes {
			if apiary.IsValidCountType(countType) {
				t.Errorf("Expected %s to be an invalid count type, but it was accepted", countType)
			}
		}
	})
}
