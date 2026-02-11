package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	apiary "github.com/chnm/apiary"
)

func TestPinkertonsActivities(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/pinkertons/activities", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.Activity
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error("Failed to unmarshal response:", err)
	}

	// Check that we got an array (even if empty)
	if data == nil {
		t.Error("Expected array of activities, got nil")
	}
}

func TestPinkertonsActivitiesWithLocations(t *testing.T) {
	// Check that we get activities with locations included
	req, _ := http.NewRequest("GET", "/pinkertons/activities?include_locations=true", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.Activity
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error("Failed to unmarshal response:", err)
	}

	// Check that we got an array
	if data == nil {
		t.Error("Expected array of activities, got nil")
	}

	// If there are activities, check that locations field exists
	if len(data) > 0 {
		// The Locations field should be initialized (even if empty)
		if data[0].Locations == nil {
			t.Error("Expected Locations field to be initialized")
		}
	}
}

func TestPinkertonsActivitiesFilterByOperative(t *testing.T) {
	// Test filtering by operative
	req, _ := http.NewRequest("GET", "/pinkertons/activities?operative=TestOperative", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var data []apiary.Activity
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error("Failed to unmarshal response:", err)
	}

	// Check that all returned activities have the specified operative
	for _, activity := range data {
		if activity.Operative.Valid && activity.Operative.String != "TestOperative" {
			t.Errorf("Expected operative 'TestOperative', got '%s'", activity.Operative.String)
		}
	}
}

func TestPinkertonsActivitiesFilterByDateRange(t *testing.T) {
	// Test filtering by date range
	req, _ := http.NewRequest("GET", "/pinkertons/activities?start_date=1900-01-01&end_date=1900-12-31", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var data []apiary.Activity
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error("Failed to unmarshal response:", err)
	}

	// Just verify we can parse the response
	if data == nil {
		t.Error("Expected array of activities, got nil")
	}
}

func TestPinkertonsActivityByID(t *testing.T) {
	// First, get all activities to find a valid ID
	req, _ := http.NewRequest("GET", "/pinkertons/activities", nil)
	response := executeRequest(req)

	if response.Code != http.StatusOK {
		t.Skip("Skipping test: no activities available")
	}

	var activities []apiary.Activity
	err := json.Unmarshal(response.Body.Bytes(), &activities)
	if err != nil || len(activities) == 0 {
		t.Skip("Skipping test: no activities available")
	}

	// Use the first activity's ID
	activityID := activities[0].ID

	// Now test getting that specific activity
	req, _ = http.NewRequest("GET", "/pinkertons/activities/"+strconv.Itoa(activityID), nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var activity apiary.Activity
	err = json.Unmarshal(response.Body.Bytes(), &activity)
	if err != nil {
		t.Error("Failed to unmarshal response:", err)
	}

	// Check that the activity has the right ID
	if activity.ID != activityID {
		t.Errorf("Expected activity ID %d, got %d", activityID, activity.ID)
	}

	// Check that locations array is initialized
	if activity.Locations == nil {
		t.Error("Expected Locations field to be initialized")
	}
}

func TestPinkertonsActivityByInvalidID(t *testing.T) {
	// Test with an invalid ID
	req, _ := http.NewRequest("GET", "/pinkertons/activities/invalid", nil)
	response := executeRequest(req)

	// Should not match the route pattern or return bad request
	if response.Code != http.StatusNotFound && response.Code != http.StatusBadRequest {
		t.Errorf("Expected 404 or 400, got %d", response.Code)
	}
}

func TestPinkertonsLocations(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/pinkertons/locations", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.Location
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error("Failed to unmarshal response:", err)
	}

	// Check that we got an array
	if data == nil {
		t.Error("Expected array of locations, got nil")
	}

	// If there are locations, verify they have the expected fields
	if len(data) > 0 {
		loc := data[0]
		if loc.ID == 0 {
			t.Error("Expected location to have an ID")
		}
	}
}

func TestPinkertonsOperatives(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/pinkertons/operatives", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []string
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error("Failed to unmarshal response:", err)
	}

	// Check that we got an array
	if data == nil {
		t.Error("Expected array of operatives, got nil")
	}
}

func TestPinkertonsSubjects(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/pinkertons/subjects", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []string
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error("Failed to unmarshal response:", err)
	}

	// Check that we got an array
	if data == nil {
		t.Error("Expected array of subjects, got nil")
	}
}

func TestPinkertonsCombinedFilters(t *testing.T) {
	// Test combining multiple filters
	req, _ := http.NewRequest("GET", "/pinkertons/activities?operative=TestOp&start_date=1900-01-01&end_date=1900-12-31&include_locations=true", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var data []apiary.Activity
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error("Failed to unmarshal response:", err)
	}

	// Verify response structure
	if data == nil {
		t.Error("Expected array of activities, got nil")
	}
}
