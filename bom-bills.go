package apiary

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
)

// ParishByYear describes a parish's canoncial name, count type, total count, start day,
// start month, end day, end month, year, week number, and week ID.
type ParishByYear struct {
	CanonicalName    string     `json:"name"`
	BillType         string     `json:"bill_type"`
	CountType        string     `json:"count_type"`
	Count            NullInt64  `json:"count"`
	StartDay         NullInt64  `json:"start_day"`
	StartMonth       NullString `json:"start_month"`
	EndDay           NullInt64  `json:"end_day"`
	EndMonth         NullString `json:"end_month"`
	Year             NullInt64  `json:"year"`
	SplitYear        string     `json:"split_year"`
	WeekNumber       int        `json:"week_number"`
	WeekID           string     `json:"week_id"`
	Missing          *bool      `json:"missing"`
	Illegible        *bool      `json:"illegible"`
	Source           NullString `json:"source"`
	UniqueIdentifier NullString `json:"unique_identifier"`
	TotalRecords     int        `json:"totalrecords"`
}

type PaginatedResponse struct {
	Data       []ParishByYear `json:"data"`
	NextCursor *string        `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

type APIParameters struct {
	StartYear  int
	EndYear    int
	Parish     []int
	BillType   string
	CountType  string
	Sort       string
	Limit      int
	Offset     int
	Page       int
	Cursor     string
	CursorYear int
	CursorWeek int
	CursorName string
}

type QueryOptions struct {
	Limit  int
	Offset int
}

// QueryBuilder holds the query string and parameters
type QueryBuilder struct {
	Query  string
	Params []interface{}
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		Params: make([]interface{}, 0),
	}
}

// AddParam adds a parameter and returns the placeholder ($1, $2, etc.)
// We do this to prevent injection problems.
func (qb *QueryBuilder) AddParam(value interface{}) string {
	qb.Params = append(qb.Params, value)
	return fmt.Sprintf("$%d", len(qb.Params))
}

// TotalBills returns to the total number of records in the database.
type TotalBills struct {
	TotalRecords NullInt64 `json:"total_records"`
}

// BillsHandler returns the bills for a given range of years. It expects a start year and
// an end year. It returns a JSON array of ParishByYear objects.
func (s *Server) BillsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse and validate query parameters
		apiParams, err := parseAPIParameters(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Printf("Error parsing API parameters: %v", err)
			return
		}

		// Build query with parameters
		qb, err := buildBillsQueryWithParams(apiParams)
		if err != nil {
			http.Error(w, "Error building query", http.StatusInternalServerError)
			log.Printf("Error building query: %v", err)
			return
		}

		// Execute query with parameters
		rows, err := s.DB.Query(context.TODO(), qb.Query, qb.Params...)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			log.Printf("Error executing query: %v", err)
			log.Printf("Query: %s", qb.Query)
			log.Printf("Params: %v", qb.Params)
			return
		}
		defer rows.Close()

		// Process results
		results := []ParishByYear{}
		for rows.Next() {
			var result ParishByYear
			err := rows.Scan(
				&result.CanonicalName,
				&result.BillType,
				&result.CountType,
				&result.Count,
				&result.StartDay,
				&result.StartMonth,
				&result.EndDay,
				&result.EndMonth,
				&result.Year,
				&result.SplitYear,
				&result.WeekNumber,
				&result.WeekID,
				&result.Missing,
				&result.Illegible,
				&result.Source,
				&result.UniqueIdentifier,
				&result.TotalRecords,
			)
			if err != nil {
				http.Error(w, "Error processing results", http.StatusInternalServerError)
				log.Printf("Error scanning row: %v", err)
				return
			}
			results = append(results, result)
		}

		if err = rows.Err(); err != nil {
			http.Error(w, "Error processing results", http.StatusInternalServerError)
			log.Printf("Error iterating through rows: %v", err)
			return
		}

		// Create paginated response
		paginatedResponse := PaginatedResponse{
			Data:    results,
			HasMore: len(results) == getEffectiveLimit(apiParams),
		}

		// Generate next cursor if there are more results
		if paginatedResponse.HasMore && len(results) > 0 {
			lastResult := results[len(results)-1]
			if nextCursor, err := generateCursor(int(lastResult.Year.Int64), int(lastResult.WeekNumber), lastResult.CanonicalName); err == nil {
				paginatedResponse.NextCursor = &nextCursor
			}
		}

		w.Header().Set("Content-Type", "application/json")
		response, err := json.Marshal(paginatedResponse)
		if err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			log.Printf("Error marshaling JSON: %v", err)
			return
		}
		w.Write(response)
	}
}

func (p *APIParameters) GetQueryOptions() QueryOptions {
	if p.Page > 0 {
		// Default to 25 items per page if using page parameter
		return QueryOptions{
			Limit:  25,
			Offset: (p.Page - 1) * 25,
		}
	}

	return QueryOptions{
		Limit:  p.Limit,
		Offset: p.Offset,
	}
}

// Helper function to parse and validate query parameters
func parseAPIParameters(r *http.Request) (APIParameters, error) {
	params := APIParameters{
		StartYear: 1648, // Default values
		EndYear:   1750,
		Parish:    []int{},
		Sort:      "year, week_number, canonical_name",
	}

	// Parse start year
	if startYear := r.URL.Query().Get("start-year"); startYear != "" {
		startYearInt, err := strconv.Atoi(startYear)
		if err != nil {
			return params, fmt.Errorf("invalid start year: %v", err)
		}
		params.StartYear = startYearInt
	}

	// Parse end year
	if endYear := r.URL.Query().Get("end-year"); endYear != "" {
		endYearInt, err := strconv.Atoi(endYear)
		if err != nil {
			return params, fmt.Errorf("invalid end year: %v", err)
		}
		params.EndYear = endYearInt
	}

	// Parse parish IDs
	if parish := r.URL.Query().Get("parish"); parish != "" {
		parishList := strings.Split(parish, ",")
		for _, p := range parishList {
			parishInt, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				return params, fmt.Errorf("invalid parish ID: %v", err)
			}
			params.Parish = append(params.Parish, parishInt)
		}
	}

	// Parse bill type
	if billType := r.URL.Query().Get("bill-type"); billType != "" {
		if !IsValidBillType(billType) {
			return params, fmt.Errorf("invalid bill type: %s", billType)
		}
		params.BillType = billType
	}

	// Parse count type
	if countType := r.URL.Query().Get("count-type"); countType != "" {
		if !IsValidCountType(countType) {
			return params, fmt.Errorf("invalid count type: %s", countType)
		}
		params.CountType = countType
	}

	// Parse pagination parameters first
	if limit := r.URL.Query().Get("limit"); limit != "" {
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			return params, fmt.Errorf("invalid limit: %v", err)
		}
		params.Limit = limitInt
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			return params, fmt.Errorf("invalid offset: %v", err)
		}
		params.Offset = offsetInt
	}

	// Handle page parameter last since it may override limit/offset
	if page := r.URL.Query().Get("page"); page != "" {
		pageInt, err := strconv.Atoi(page)
		if err != nil {
			return params, fmt.Errorf("invalid page number: %v", err)
		}
		params.Page = pageInt
	}

	// Parse sorting
	if sort := r.URL.Query().Get("sort"); sort != "" {
		params.Sort = sort
	}

	// Parse cursor
	if cursor := r.URL.Query().Get("cursor"); cursor != "" {
		params.Cursor = cursor
		if year, week, name, err := parseCursor(cursor); err == nil {
			params.CursorYear = year
			params.CursorWeek = week
			params.CursorName = name
		} else {
			return params, fmt.Errorf("invalid cursor: %v", err)
		}
	}

	return params, nil
}

// Helper function to build the SQL query with parameters
func buildBillsQueryWithParams(params APIParameters) (*QueryBuilder, error) {
	qb := NewQueryBuilder()

	// For cursor-based pagination, skip the expensive COUNT(*) OVER() calculation
	// Only calculate total count for first page or legacy pagination
	var selectClause string
	if params.Cursor != "" {
		// Fast cursor query without total count for subsequent pages
		selectClause = `
    SELECT
        p.canonical_name,
        b.bill_type,
        b.count_type,
        b.count,
        w.start_day,
        w.start_month,
        w.end_day,
        w.end_month,
        b.year,
        w.split_year,
        w.week_number,
        b.week_id,
        b.missing,
        b.illegible,
        b.source,
        b.unique_identifier,
        0 AS totalrecords`
	} else {
		// Include total count for first page and legacy pagination
		selectClause = `
    SELECT
        p.canonical_name,
        b.bill_type,
        b.count_type,
        b.count,
        w.start_day,
        w.start_month,
        w.end_day,
        w.end_month,
        b.year,
        w.split_year,
        w.week_number,
        b.week_id,
        b.missing,
        b.illegible,
        b.source,
        b.unique_identifier,
        COUNT(*) OVER() AS totalrecords`
	}

	baseQuery := selectClause + `
    FROM
        bom.bill_of_mortality b
    JOIN
        bom.parishes p ON p.id = b.parish_id
    JOIN
        bom.year y ON y.year = b.year
    JOIN
        bom.week w ON w.joinid = b.week_id
    WHERE 1=1`

	// Build WHERE clause
	var conditions []string

	if params.StartYear != 0 {
		conditions = append(conditions, fmt.Sprintf("b.year >= %s", qb.AddParam(params.StartYear)))
	}

	if params.EndYear != 0 {
		conditions = append(conditions, fmt.Sprintf("b.year <= %s", qb.AddParam(params.EndYear)))
	}

	if len(params.Parish) > 0 {
		// Convert []int to []interface{} for the parameter
		parishParams := make([]string, len(params.Parish))
		for i, p := range params.Parish {
			parishParams[i] = qb.AddParam(p)
		}
		conditions = append(conditions, fmt.Sprintf("b.parish_id IN (%s)", strings.Join(parishParams, ",")))
	}

	if params.BillType != "" {
		conditions = append(conditions, fmt.Sprintf("b.bill_type = %s", qb.AddParam(params.BillType)))
	}

	if params.CountType != "" {
		conditions = append(conditions, fmt.Sprintf("b.count_type = %s", qb.AddParam(params.CountType)))
	}

	// Handle cursor-based pagination
	if params.Cursor != "" {
		yearParam := qb.AddParam(params.CursorYear)
		weekParam := qb.AddParam(params.CursorWeek)
		nameParam := qb.AddParam(params.CursorName)
		conditions = append(conditions, fmt.Sprintf("(b.year, w.week_number, p.canonical_name) > (%s, %s, %s)", yearParam, weekParam, nameParam))
	}

	// Add conditions to query
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Add sorting (ensure consistent ordering for cursor pagination)
	if params.Sort != "" {
		baseQuery += " ORDER BY " + params.Sort + " ASC"
	} else {
		baseQuery += " ORDER BY b.year, w.week_number, p.canonical_name ASC"
	}

	// Handle pagination - cursor-based is the default and preferred method
	if params.Cursor != "" {
		// Cursor pagination (preferred for performance)
		limit := getEffectiveLimit(params)
		baseQuery += fmt.Sprintf(" LIMIT %s", qb.AddParam(limit))
	} else if params.Page > 0 {
		// Legacy page-based pagination for compatibility
		limit := 100
		offset := (params.Page - 1) * limit
		baseQuery += fmt.Sprintf(" LIMIT %s OFFSET %s", qb.AddParam(limit), qb.AddParam(offset))
	} else if params.Limit > 0 {
		// Direct limit/offset pagination
		baseQuery += fmt.Sprintf(" LIMIT %s", qb.AddParam(params.Limit))
		if params.Offset > 0 {
			baseQuery += fmt.Sprintf(" OFFSET %s", qb.AddParam(params.Offset))
		}
	} else {
		// Default to cursor-based pagination
		defaultLimit := 100
		baseQuery += fmt.Sprintf(" LIMIT %s", qb.AddParam(defaultLimit))
	}

	qb.Query = baseQuery
	return qb, nil
}

// IsValidBillType checks if the provided bill type is valid
func IsValidBillType(billType string) bool {
	validTypes := map[string]bool{
		"weekly":  true,
		"general": true,
		"total":   true,
	}
	return validTypes[billType]
}

// IsValidCountType checks if the provided count type is valid
func IsValidCountType(countType string) bool {
	validTypes := map[string]bool{
		"buried": true,
		"plague": true,
	}
	return validTypes[countType]
}

// generateCursor creates a base64-encoded cursor from year, week, and name
func generateCursor(year, week int, name string) (string, error) {
	cursorData := fmt.Sprintf("%d|%d|%s", year, week, name)
	return base64.URLEncoding.EncodeToString([]byte(cursorData)), nil
}

// parseCursor decodes a base64 cursor back to year, week, and name
func parseCursor(cursor string) (int, int, string, error) {
	decoded, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, 0, "", err
	}

	parts := strings.Split(string(decoded), "|")
	if len(parts) != 3 {
		return 0, 0, "", fmt.Errorf("invalid cursor format")
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, "", fmt.Errorf("invalid year in cursor: %v", err)
	}

	week, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, "", fmt.Errorf("invalid week in cursor: %v", err)
	}

	return year, week, parts[2], nil
}

// getEffectiveLimit returns the appropriate limit for the query
func getEffectiveLimit(params APIParameters) int {
	if params.Cursor != "" {
		return 100 // Default page size for cursor pagination
	}
	if params.Page > 0 {
		return 100
	}
	if params.Limit > 0 {
		return params.Limit
	}
	return 100
}

// TotalBillsHandler returns the total number of bills in the database.
// This number is required for pagination in the web application.
func (s *Server) TotalBillsHandler() http.HandlerFunc {
	queryWeekly := `
	SELECT
		COUNT(*)
	FROM
		bom.bill_of_mortality
	WHERE 
		bill_type = 'weekly';
	`

	queryGeneral := `
	SELECT
		COUNT(*)
	FROM
		bom.bill_of_mortality
	WHERE	
		bill_type = 'general';
	`

	queryChristenings := `
	SELECT
		COUNT(*)
	FROM	
		bom.christenings;
	`

	queryCauses := `
	SELECT
		COUNT(*)
	FROM 
		bom.causes_of_death;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		totalValues := r.URL.Query().Get("type")

		if totalValues == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		results := make([]TotalBills, 0)
		var row TotalBills
		var rows pgx.Rows
		var err error

		switch {
		case totalValues == "weekly":
			rows, err = s.DB.Query(context.TODO(), queryWeekly)
		case totalValues == "general":
			rows, err = s.DB.Query(context.TODO(), queryGeneral)
		case totalValues == "christenings":
			rows, err = s.DB.Query(context.TODO(), queryChristenings)
		case totalValues == "causes":
			rows, err = s.DB.Query(context.TODO(), queryCauses)
		}
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&row.TotalRecords)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			results = append(results, row)
		}

		response, _ := json.Marshal(results)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}
}

// Statistics
type YearlySummary struct {
	Year           int `json:"year"`
	WeeksCompleted int `json:"weeksCompleted"`
	RowsCount      int `json:"rowsCount"`
	TotalCount     int `json:"totalCount"`
}

type WeeklySummary struct {
	Year       int `json:"year"`
	WeekNumber int `json:"weekNumber"`
	RowsCount  int `json:"rowsCount"`
}

// ParishYearlySummary represents total counts by parish and year for small multiple visualizations
type ParishYearlySummary struct {
	Year        int    `json:"year"`
	ParishName  string `json:"parish_name"`
	TotalBuried int    `json:"total_buried"`
	TotalPlague *int   `json:"total_plague"`
}

func buildYearlyStatsQuery() string {
	query := `
  WITH year_range AS (
        SELECT generate_series(1636, 1754) AS year
    ),
    weekly_stats AS (
        SELECT 
            b.year as year,
            COUNT(DISTINCT b.week_id) as weeks_completed,
            COUNT(*) as rows_count
        FROM bom.bill_of_mortality b
        WHERE b.bill_type = 'weekly'
        GROUP BY b.year
    )
    SELECT 
        yr.year,
        COALESCE(ws.weeks_completed, 0) as weeks_completed,
        COALESCE(ws.rows_count, 0) as rows_count,
        53 as total_count
    FROM year_range yr
    LEFT JOIN weekly_stats ws ON yr.year = ws.year
    ORDER BY yr.year;
    `
	return query
}

func buildWeeklyStatsQuery() string {
	query := `
    WITH year_week_range AS (
        SELECT 
            y.year,
            w.number as week_number
        FROM generate_series(1636, 1754) y(year)
        CROSS JOIN generate_series(1, 53) w(number)
    ),
    weekly_stats AS (
        SELECT 
            b.year as year,
            w.week_number,
            COUNT(*) as rows_count
        FROM bom.bill_of_mortality b
        JOIN bom.week w ON w.joinid = b.week_id
        WHERE b.bill_type = 'weekly'
        GROUP BY b.year, w.week_number
    )
    SELECT 
        yr.year,
        yr.week_number,
        COALESCE(ws.rows_count, 0) as rows_count
    FROM year_week_range yr
    LEFT JOIN weekly_stats ws ON yr.year = ws.year AND yr.week_number = ws.week_number
    ORDER BY yr.year, yr.week_number;
    `
	return query
}

func buildParishYearlyStatsQuery(parishName string) (*QueryBuilder, error) {
	qb := NewQueryBuilder()

	query := `
    SELECT 
        b.year,
        p.canonical_name as parish_name,
        SUM(CASE WHEN b.count_type = 'Buried' THEN COALESCE(b.count, 0) ELSE 0 END) as total_buried,
        NULLIF(SUM(CASE WHEN b.count_type = 'Plague' THEN COALESCE(b.count, 0) ELSE 0 END), 0) as total_plague
    FROM bom.bill_of_mortality b
    JOIN bom.parishes p ON p.id = b.parish_id
    WHERE b.bill_type = 'weekly'`

	if parishName != "" {
		query += fmt.Sprintf(" AND p.canonical_name = %s", qb.AddParam(parishName))
	}

	query += `
    GROUP BY b.year, p.canonical_name
    ORDER BY p.canonical_name, b.year`

	qb.Query = query
	return qb, nil
}

func (s *Server) StatisticsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statType := r.URL.Query().Get("type")
		parishName := r.URL.Query().Get("parish")

		var rows pgx.Rows
		var err error

		switch statType {
		case "weekly":
			query := buildWeeklyStatsQuery()
			rows, err = s.DB.Query(context.TODO(), query)
		case "yearly":
			query := buildYearlyStatsQuery()
			rows, err = s.DB.Query(context.TODO(), query)
		case "parish-yearly":
			qb, buildErr := buildParishYearlyStatsQuery(parishName)
			if buildErr != nil {
				http.Error(w, "Error building query", http.StatusInternalServerError)
				log.Printf("Error building parish-yearly query: %v", buildErr)
				return
			}
			rows, err = s.DB.Query(context.TODO(), qb.Query, qb.Params...)
		default:
			http.Error(w, "Invalid type parameter. Must be 'weekly', 'yearly', or 'parish-yearly'", http.StatusBadRequest)
			return
		}
		if err != nil {
			log.Printf("Database error executing query: %v", err)
			if statType == "parish-yearly" {
				log.Printf("Parish name: %s", parishName)
			}
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		switch statType {
		case "weekly":
			stats := []WeeklySummary{}
			for rows.Next() {
				var summary WeeklySummary
				err := rows.Scan(&summary.Year, &summary.WeekNumber, &summary.RowsCount)
				if err != nil {
					log.Printf("Error scanning weekly summary: %v", err)
					http.Error(w, fmt.Sprintf("Error processing results: %v", err), http.StatusInternalServerError)
					return
				}
				stats = append(stats, summary)
			}

			// Check for errors from iterating over rows
			if err = rows.Err(); err != nil {
				log.Printf("Error iterating over rows: %v", err)
				http.Error(w, fmt.Sprintf("Error processing results: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(stats); err != nil {
				log.Printf("Error encoding JSON response: %v", err)
			}

		case "yearly":
			stats := []YearlySummary{}
			for rows.Next() {
				var summary YearlySummary
				err := rows.Scan(&summary.Year, &summary.WeeksCompleted,
					&summary.RowsCount, &summary.TotalCount)
				if err != nil {
					log.Printf("Error scanning yearly summary: %v", err)
					http.Error(w, fmt.Sprintf("Error processing results: %v", err), http.StatusInternalServerError)
					return
				}
				stats = append(stats, summary)
			}

			// Check for errors from iterating over rows
			if err = rows.Err(); err != nil {
				log.Printf("Error iterating over rows: %v", err)
				http.Error(w, fmt.Sprintf("Error processing results: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(stats); err != nil {
				log.Printf("Error encoding JSON response: %v", err)
			}

		case "parish-yearly":
			stats := []ParishYearlySummary{}
			for rows.Next() {
				var summary ParishYearlySummary
				err := rows.Scan(
					&summary.Year,
					&summary.ParishName,
					&summary.TotalBuried,
					&summary.TotalPlague,
				)
				if err != nil {
					log.Printf("Error scanning parish-yearly summary: %v", err)
					http.Error(w, fmt.Sprintf("Error processing results: %v", err), http.StatusInternalServerError)
					return
				}
				stats = append(stats, summary)
			}

			// Check for errors from iterating over rows
			if err = rows.Err(); err != nil {
				log.Printf("Error iterating over rows: %v", err)
				http.Error(w, fmt.Sprintf("Error processing results: %v", err), http.StatusInternalServerError)
				return
			}

			log.Printf("Returning %d parish-yearly summary records", len(stats))
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(stats); err != nil {
				log.Printf("Error encoding JSON response: %v", err)
			}
		}
	}
}
