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

type DeathsAPIParameters struct {
	StartYear int
	EndYear   int
	Death     []string
	Sort      string
}

// DeathCauses returns a list of causes of death with a count of deaths for each
// cause and related metadata.
type DeathCauses struct {
	Death            string     `json:"death"`
	BillType         string     `json:"bill_type"`
	Count            NullInt64  `json:"count"`
	Definition       NullString `json:"definition"`
	DefinitionSource NullString `json:"definition_source"`
	WeekID           string     `json:"week_id"`
	WeekNumber       NullInt64  `json:"week_number"`
	StartDay         NullInt64  `json:"start_day"`
	StartMonth       NullString `json:"start_month"`
	EndDay           NullInt64  `json:"end_day"`
	EndMonth         NullString `json:"end_month"`
	Year             NullInt64  `json:"year"`
	SplitYear        NullString `json:"split_year"`
	TotalRecords     int        `json:"totalrecords"`
}

// Causes describes a cause of death.
type Causes struct {
	Name string `json:"name"`
}

// DeathCausesHandler returns a JSON array of causes of death. The list of causes
// depends on whether a user has provided a comma-separated list of causes. If
// no list is provided, it returns the entire list of causes.
func (s *Server) DeathCausesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startYear := r.URL.Query().Get("start-year")
		endYear := r.URL.Query().Get("end-year")
		causes := r.URL.Query().Get("id")
		billType := r.URL.Query().Get("bill-type")
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

		apiParams := DeathsAPIParameters{
			StartYear: 1648,
			EndYear:   1750,
			Death:     []string{},
			Sort:      "year, week_number, death",
		}

		// If a start year is provided, update the API parameters
		if startYear != "" {
			startYearInt, err := strconv.Atoi(startYear)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				log.Println("start year is not an integer", err)
				return
			}

			apiParams.StartYear = startYearInt
		}

		// If an end year is provided, update the API parameters
		if endYear != "" {
			endYearInt, err := strconv.Atoi(endYear)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				log.Println("end year is not an integer", err)
				return
			}

			apiParams.EndYear = endYearInt
		}

		// if a cause string is provided, update the API parameters
		if causes != "" {
			causesList := strings.Split(causes, ",")
			var causesStr []string

			for _, p := range causesList {
				causeStr := strings.TrimSpace(p)
				if causeStr == "" {
					http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
					log.Println("cause is an empty string")
					return
				}
				causesStr = append(causesStr, causeStr)
			}

			apiParams.Death = causesStr
		}

		// Validate bill type if provided
		if billType != "" {
			if !IsValidBillType(billType) {
				http.Error(w, "Invalid bill type", http.StatusBadRequest)
				log.Printf("Invalid bill type: %s", billType)
				return
			}
		}

		query := `
    SELECT 
        c.death,
        c.bill_type,
        c.count, 
        c.definition,
        c.definition_source,
        c.week_id,
        w.week_number,
        w.start_day, 
        w.start_month, 
        w.end_day, 
        w.end_month, 
        y.year,
        w.split_year,
        COUNT(*) OVER() AS totalrecords
    FROM 
        bom.causes_of_death c
    JOIN 
        bom.week w ON w.joinid = c.week_id
    JOIN
        bom.year y ON y.year = w.year
    WHERE 
        y.year::int >= $1
        AND y.year::int <= $2
        AND count IS NOT NULL
    `

		paramCount := 2

		if len(apiParams.Death) > 0 {
			paramCount++
			query += fmt.Sprintf(" AND c.death = ANY($%d)", paramCount)
		}

		if billType != "" {
			paramCount++
			query += fmt.Sprintf(" AND c.bill_type = $%d", paramCount)
		}

		query += " ORDER BY " + apiParams.Sort

		if limit != "" {
			limitInt, err := strconv.Atoi(limit)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				log.Println("limit is not an integer", err)
				return
			}

			query += " LIMIT " + strconv.Itoa(limitInt)
		}

		// If an offset is provided, add it to the query
		if offset != "" {
			offsetInt, err := strconv.Atoi(offset)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				log.Println("offset is not an integer", err)
				return
			}

			query += " OFFSET " + strconv.Itoa(offsetInt)
		}

		results := make([]DeathCauses, 0)
		var row DeathCauses
		var rows pgx.Rows
		var err error

		// Build parameters slice
		params := []interface{}{apiParams.StartYear, apiParams.EndYear}
		
		if len(apiParams.Death) > 0 {
			params = append(params, apiParams.Death)
		}
		
		if billType != "" {
			params = append(params, billType)
		}

		rows, err = s.DB.Query(context.TODO(), query, params...)

		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Fatal("Error preparing statement", err)
			return
		}

		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(
				&row.Death,
				&row.BillType,
				&row.Count,
				&row.Definition,
				&row.DefinitionSource,
				&row.WeekID,
				&row.WeekNumber,
				&row.StartDay,
				&row.StartMonth,
				&row.EndDay,
				&row.EndMonth,
				&row.Year,
				&row.SplitYear,
				&row.TotalRecords,
			)
			if err != nil {
				log.Printf("Error scanning row: %v", err)
				log.Printf("Types: death=%T, billType=%T, count=%T, definition=%T, definitionSource=%T, weekID=%T, weekNumber=%T, startDay=%T, startMonth=%T, endDay=%T, endMonth=%T, year=%T, splitYear=%T, totalRecords=%T",
					row.Death, row.BillType, row.Count, row.Definition, row.DefinitionSource, row.WeekID, row.WeekNumber, row.StartDay, row.StartMonth, row.EndDay, row.EndMonth, row.Year, row.SplitYear, row.TotalRecords)
				continue
			}
			results = append(results, row)
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, _ := json.Marshal(results)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}
}


func (s *Server) ListCausesHandler() http.HandlerFunc {
	// Query to get a unique list of causes of death

	query := `
	SELECT DISTINCT
		death
	FROM 
		bom.causes_of_death
	ORDER BY 
		death ASC
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]Causes, 0)
		var row Causes

		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Name)
			if err != nil {
				log.Println(err)
			}
			results = append(results, row)
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		response, _ := json.Marshal(results)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}
}
