package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	apiary "github.com/chnm/apiary"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pinkertonsQueryCounter struct {
	count atomic.Int64
}

func (counter *pinkertonsQueryCounter) TraceQueryStart(
	ctx context.Context,
	_ *pgx.Conn,
	_ pgx.TraceQueryStartData,
) context.Context {
	counter.count.Add(1)
	return ctx
}

func (*pinkertonsQueryCounter) TraceQueryEnd(
	context.Context,
	*pgx.Conn,
	pgx.TraceQueryEndData,
) {
}

func fetchDetectiveActivities(t *testing.T, path string) []apiary.Activity {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		t.Fatalf("Create request: %v", err)
	}
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var activities []apiary.Activity
	if err := json.Unmarshal(response.Body.Bytes(), &activities); err != nil {
		t.Fatalf("Unmarshal response: %v", err)
	}
	return activities
}

func TestDetectivesActivities(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/pinkertons/activities?limit=10", nil)
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

func TestDetectivesActivitiesDefaultLimit(t *testing.T) {
	const defaultLimit = 500

	defaultPage := fetchDetectiveActivities(t, "/pinkertons/activities")
	if len(defaultPage) != defaultLimit {
		t.Fatalf(
			"Default page length = %d, want %d",
			len(defaultPage),
			defaultLimit,
		)
	}

	explicitPage := fetchDetectiveActivities(
		t,
		"/pinkertons/activities?limit=501",
	)
	if len(explicitPage) != defaultLimit+1 {
		t.Fatalf(
			"Explicit page length = %d, want %d",
			len(explicitPage),
			defaultLimit+1,
		)
	}

	for index := range defaultPage {
		if defaultPage[index].ID != explicitPage[index].ID {
			t.Fatalf(
				"Default page activity %d ID = %d, want %d",
				index,
				defaultPage[index].ID,
				explicitPage[index].ID,
			)
		}
	}
}

func TestDetectivesActivitiesUsesTwoQueries(t *testing.T) {
	config := s.DB.Config()
	queryCounter := &pinkertonsQueryCounter{}
	config.ConnConfig.Tracer = queryCounter

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Fatalf("Failed to create traced database pool: %v", err)
	}
	t.Cleanup(pool.Close)

	server := &apiary.Server{DB: pool}
	req := httptest.NewRequest(
		http.MethodGet,
		"/pinkertons/activities?limit=10",
		nil,
	)
	response := httptest.NewRecorder()
	server.ActivitiesHandler().ServeHTTP(response, req)
	checkResponseCode(t, http.StatusOK, response.Code)

	if got := queryCounter.count.Load(); got != 2 {
		t.Fatalf("Expected 2 database queries, got %d", got)
	}
}

func TestDetectivesActivitiesOffsetPagination(t *testing.T) {
	firstTwo := fetchDetectiveActivities(
		t,
		"/pinkertons/activities?limit=2&offset=0",
	)
	if len(firstTwo) != 2 {
		t.Fatalf("First page length = %d, want 2", len(firstTwo))
	}

	secondActivity := fetchDetectiveActivities(
		t,
		"/pinkertons/activities?limit=1&offset=1",
	)
	if len(secondActivity) != 1 {
		t.Fatalf("Second page length = %d, want 1", len(secondActivity))
	}
	if secondActivity[0].ID != firstTwo[1].ID {
		t.Fatalf(
			"Offset activity ID = %d, want %d",
			secondActivity[0].ID,
			firstTwo[1].ID,
		)
	}

	var locationID int
	for _, activity := range firstTwo {
		if len(activity.Locations) > 0 {
			locationID = activity.Locations[0].ID
			break
		}
	}
	if locationID == 0 {
		t.Fatal("First page has no location to test filtered pagination")
	}

	filteredPage := fetchDetectiveActivities(
		t,
		"/pinkertons/activities?location_id="+
			strconv.Itoa(locationID)+
			"&limit=1&offset=0",
	)
	if len(filteredPage) != 1 {
		t.Fatalf("Filtered page length = %d, want 1", len(filteredPage))
	}
}

func TestDetectivesActivitiesWithLocations(t *testing.T) {
	// Check that we get activities with locations included
	req, _ := http.NewRequest("GET", "/pinkertons/activities?include_locations=true&limit=10", nil)
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

func TestDetectivesActivitiesFilterByOperative(t *testing.T) {
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

func TestDetectivesActivitiesFilterByDateRange(t *testing.T) {
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

func TestDetectivesActivityByID(t *testing.T) {
	// First, get all activities to find a valid ID
	req, _ := http.NewRequest("GET", "/pinkertons/activities?limit=1", nil)
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

func TestDetectivesActivityByInvalidID(t *testing.T) {
	// Test with an invalid ID
	req, _ := http.NewRequest("GET", "/pinkertons/activities/invalid", nil)
	response := executeRequest(req)

	// Should not match the route pattern or return bad request
	if response.Code != http.StatusNotFound && response.Code != http.StatusBadRequest {
		t.Errorf("Expected 404 or 400, got %d", response.Code)
	}
}

func TestDetectivesLocations(t *testing.T) {
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

func TestDetectivesOperatives(t *testing.T) {
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

func TestDetectivesSubjects(t *testing.T) {
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

func TestDetectivesCombinedFilters(t *testing.T) {
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
