package apiary

import (
	"context"
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
	CanonicalName string     `json:"name"`
	BillType      string     `json:"bill_type"`
	CountType     string     `json:"count_type"`
	Count         NullInt64  `json:"count"`
	StartDay      NullInt64  `json:"start_day"`
	StartMonth    NullString `json:"start_month"`
	EndDay        NullInt64  `json:"end_day"`
	EndMonth      NullString `json:"end_month"`
	Year          NullInt64  `json:"year"`
	SplitYear     string     `json:"split_year"`
	WeekNo        int        `json:"week_no"`
	WeekID        string     `json:"week_id"`
	Missing       *bool      `json:"missing"`
	Illegible     *bool      `json:"illegible"`
	Source        NullString `json:"source"`
	TotalRecords  int        `json:"totalrecords"`
}

type APIParameters struct {
	StartYear int
	EndYear   int
	Parish    []int
	BillType  string
	CountType string
	Sort      string
	Limit     int
	Offset    int
	Page      int
}

type QueryOptions struct {
	Limit  int
	Offset int
}

// TotalBills returns to the total number of records in the database. We need this
// number to get pagination working.
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

		// Build query
		query, err := buildBillsQuery(apiParams)
		if err != nil {
			http.Error(w, "Error building query", http.StatusInternalServerError)
			log.Printf("Error building query: %v", err)
			return
		}

		// Execute query
		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			log.Printf("Error executing query: %v", err)
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
				&result.WeekNo,
				&result.WeekID,
				&result.Missing,
				&result.Illegible,
				&result.Source,
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

		// Return results (empty array if no results found)
		w.Header().Set("Content-Type", "application/json")
		response, err := json.Marshal(results)
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
		Sort:      "year, week_no, canonical_name",
	}

	// Log raw query parameters for debugging
	log.Printf("Raw query parameters: %v", r.URL.Query())

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
		if !isValidBillType(billType) {
			return params, fmt.Errorf("invalid bill type: %s", billType)
		}
		params.BillType = billType
	}

	// Parse count type
	if countType := r.URL.Query().Get("count-type"); countType != "" {
		if !isValidCountType(countType) {
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
		log.Printf("Setting explicit limit to: %d", limitInt)
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			return params, fmt.Errorf("invalid offset: %v", err)
		}
		params.Offset = offsetInt
		log.Printf("Setting offset to: %d", offsetInt)
	}

	// Handle page parameter last since it may override limit/offset
	if page := r.URL.Query().Get("page"); page != "" {
		pageInt, err := strconv.Atoi(page)
		if err != nil {
			return params, fmt.Errorf("invalid page number: %v", err)
		}
		params.Page = pageInt
		log.Printf("Using page number: %d", pageInt)
	}

	// Parse sorting
	if sort := r.URL.Query().Get("sort"); sort != "" {
		params.Sort = sort
	}

	// Log final parameter values
	log.Printf("Final API parameters: %+v", params)

	return params, nil
}

// Helper function to build the SQL query
func buildBillsQuery(params APIParameters) (string, error) {
	baseQuery := `
    SELECT
        p.canonical_name,
        b.bill_type,
        b.count_type,
        b.count,
        w.start_day,
        w.start_month,
        w.end_day,
        w.end_month,
        y.year,
        w.split_year,
        w.week_no,
        b.week_id,
        b.missing,
        b.illegible,
        b.source,
        COUNT(*) OVER() AS totalrecords
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
		conditions = append(conditions, fmt.Sprintf("b.year >= %d", params.StartYear))
	}
	if params.EndYear != 0 {
		conditions = append(conditions, fmt.Sprintf("b.year <= %d", params.EndYear))
	}
	if len(params.Parish) > 0 {
		conditions = append(conditions, fmt.Sprintf("b.parish_id IN (%s)",
			strings.Trim(strings.Join(strings.Fields(fmt.Sprint(params.Parish)), ","), "[]")))
	}
	if params.BillType != "" {
		conditions = append(conditions, fmt.Sprintf("b.bill_type = '%s'", params.BillType))
	}
	if params.CountType != "" {
		conditions = append(conditions, fmt.Sprintf("b.count_type = '%s'", params.CountType))
	}

	// Add conditions to query
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Add sorting
	if params.Sort != "" {
		baseQuery += " ORDER BY " + params.Sort + " ASC"
	}

	// Handle pagination
	if params.Page > 0 {
		limit := 25
		offset := (params.Page - 1) * limit
		baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	} else if params.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT %d", params.Limit)
		if params.Offset > 0 {
			baseQuery += fmt.Sprintf(" OFFSET %d", params.Offset)
		}
	}

	log.Printf("Final query with pagination: %s", baseQuery)
	return baseQuery, nil
}

// Helper function to validate bill types
func isValidBillType(billType string) bool {
	validTypes := map[string]bool{
		"Weekly":  true,
		"General": true,
		"Total":   true,
	}
	return validTypes[billType]
}

// Helper function to validate count types
func isValidCountType(countType string) bool {
	validTypes := map[string]bool{
		"Buried": true,
		"Plague": true,
	}
	return validTypes[countType]
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
		bill_type = 'Weekly';
	`

	queryGeneral := `
	SELECT
		COUNT(*)
	FROM
		bom.bill_of_mortality
	WHERE	
		bill_type = 'General';
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
		case totalValues == "Weekly":
			rows, err = s.DB.Query(context.TODO(), queryWeekly)
		case totalValues == "General":
			rows, err = s.DB.Query(context.TODO(), queryGeneral)
		case totalValues == "Christenings":
			rows, err = s.DB.Query(context.TODO(), queryChristenings)
		case totalValues == "Causes":
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
	Year      int `json:"year"`
	WeekNo    int `json:"weekNo"`
	RowsCount int `json:"rowsCount"`
}

func buildYearlyStatsQuery() string {
	return `
  WITH year_range AS (
        SELECT generate_series(1636, 1754) AS year
    ),
    weekly_stats AS (
        SELECT 
            b.year as year,
            COUNT(DISTINCT b.week_id) as weeks_completed,
            COUNT(*) as rows_count
        FROM bom.bill_of_mortality b
        WHERE b.bill_type = 'Weekly'
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
}

func buildWeeklyStatsQuery() string {
	return `
    WITH year_week_range AS (
        SELECT 
            y.year,
            w.number as week_no
        FROM generate_series(1636, 1754) y(year)
        CROSS JOIN generate_series(1, 53) w(number)
    ),
    weekly_stats AS (
        SELECT 
            b.year as year,
            w.week_no,
            COUNT(*) as rows_count
        FROM bom.bill_of_mortality b
        JOIN bom.week w ON w.joinid = b.week_id
        WHERE b.bill_type = 'Weekly'
        GROUP BY b.year, w.week_no
    )
    SELECT 
        yr.year,
        yr.week_no,
        COALESCE(ws.rows_count, 0) as rows_count
    FROM year_week_range yr
    LEFT JOIN weekly_stats ws ON yr.year = ws.year AND yr.week_no = ws.week_no
    ORDER BY yr.year, yr.week_no;
    `
}

func (s *Server) StatisticsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statType := r.URL.Query().Get("type")

		var query string
		switch statType {
		case "weekly":
			query = buildWeeklyStatsQuery()
		case "yearly":
			query = buildYearlyStatsQuery()
		default:
			http.Error(w, "Invalid type parameter. Must be 'weekly' or 'yearly'", http.StatusBadRequest)
			return
		}

		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		if statType == "weekly" {
			stats := []WeeklySummary{}
			for rows.Next() {
				var summary WeeklySummary
				err := rows.Scan(&summary.Year, &summary.WeekNo, &summary.RowsCount)
				if err != nil {
					http.Error(w, "Error processing results", http.StatusInternalServerError)
					return
				}
				stats = append(stats, summary)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(stats)
			return
		}

		stats := []YearlySummary{}
		for rows.Next() {
			var summary YearlySummary
			err := rows.Scan(&summary.Year, &summary.WeeksCompleted,
				&summary.RowsCount, &summary.TotalCount)
			if err != nil {
				http.Error(w, "Error processing results", http.StatusInternalServerError)
				return
			}
			stats = append(stats, summary)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}
