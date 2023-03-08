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

// ChristeningsByYear describes a christening's description, total count, week number,
// week ID, and year.
type ChristeningsByYear struct {
	Christening string    `json:"christening"`
	TotalCount  NullInt64 `json:"count"`
	WeekNo      int       `json:"week_no"`
	StartDay    NullInt64 `json:"start_day"`
	StartMonth  string    `json:"start_month"`
	EndDay      NullInt64 `json:"end_day"`
	EndMonth    string    `json:"end_month"`
	Year        int       `json:"year"`
	SplitYear   string    `json:"split_year"`
	// LocID       int       `json:"loc_id"`
}

// Christenings describes a christening location.
type Christenings struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// ChristeningsHandler returns the christenings for a given range of years. It expects a start year and
// end year as query parameters.
func (s *Server) ChristeningsHandler() http.HandlerFunc {

	queryLocation := `
	SELECT
		c.christening,
		c.count,
		c.week_number,
		c.start_day,
		c.start_month,
		c.end_day,
		c.end_month,
		y.year
	FROM
		bom.christenings c
	JOIN
		bom.year y ON y.year = c.year
	JOIN
		bom.christening_locations l ON l.name = c.christening
	WHERE
		y.year >= $1
		AND y.year < $2
		AND (
			$3::int[] IS NULL
			OR l.id = ANY($3::int[])
		)	
	ORDER BY
		year ASC,
		week_number ASC
	LIMIT $4
	OFFSET $5;
	`

	query := `
	SELECT
		c.christening,
		c.count,
		c.week_number,
		c.start_day,
		c.start_month,
		c.end_day,
		c.end_month,
		y.year
	FROM
		bom.christenings c
	JOIN
		bom.year y ON y.year = c.year
	WHERE
		y.year >= $1
		AND y.year < $2
	ORDER BY
		year ASC,
		week_number ASC
	LIMIT $3
	OFFSET $4;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		startYear := r.URL.Query().Get("start-year")
		endYear := r.URL.Query().Get("end-year")
		location := r.URL.Query().Get("id")
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

		if startYear == "" || endYear == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		startYearInt, err := strconv.Atoi(startYear)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		endYearInt, err := strconv.Atoi(endYear)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if limit == "" {
			limit = "25"
		}
		if offset == "" {
			offset = "0"
		}

		// location needs to be a postgres array
		location = fmt.Sprintf("{%s}", strings.TrimSpace(location))

		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		results := make([]ChristeningsByYear, 0)
		var row ChristeningsByYear
		var rows pgx.Rows

		switch {
		case location == "{}":
			rows, err = s.DB.Query(context.TODO(), query, startYearInt, endYearInt, limitInt, offsetInt)
		case location != "{}":
			rows, err = s.DB.Query(context.TODO(), queryLocation, startYearInt, endYearInt, location, limitInt, offsetInt)
		default:
			rows, err = s.DB.Query(context.TODO(), query, startYearInt, endYearInt, limitInt, offsetInt)
		}

		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(
				&row.Christening,
				&row.TotalCount,
				&row.WeekNo,
				&row.StartDay,
				&row.StartMonth,
				&row.EndDay,
				&row.EndMonth,
				&row.Year,
				// &row.SplitYear,
				// &row.LocID,
			)
			if err != nil {
				log.Println(err)
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

func (s *Server) ListChristeningsHandler() http.HandlerFunc {

	query := `
	SELECT DISTINCT
		name,
		id
	FROM 
		bom.christening_locations
	ORDER BY 
		name ASC
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]Christenings, 0)
		var row Christenings

		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Name, &row.ID)
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
