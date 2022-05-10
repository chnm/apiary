package apiary

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// ChristeningsByYear describes a christening's description, total count, week number,
// week ID, and year.
type ChristeningsByYear struct {
	ChristeningsDesc string    `json:"christenings_desc"`
	TotalCount       NullInt64 `json:"count"`
	WeekNo           int       `json:"week_no"`
	WeekID           string    `json:"week_id"`
	Year             int       `json:"year"`
}

// ChristeningsHandler returns the christenings for a given range of years. It expects a start year and
// end year as query parameters.
func (s *Server) ChristeningsHandler() http.HandlerFunc {

	query := `
	SELECT
		c.christening_desc,
		c.count,
		w.week_no,
		c.week_id,
		y.year
	FROM
		bom.christenings c
	JOIN
		bom.year y ON y.year_id = c.year_id
	JOIN
		bom.week w ON w.week_id = c.week_id
	WHERE
		year >= $1
		AND year < $2
	ORDER BY
		count;
	`

	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["christenings"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {
		startYear := r.URL.Query().Get("startYear")
		endYear := r.URL.Query().Get("endYear")

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

		results := make([]ChristeningsByYear, 0)
		var row ChristeningsByYear

		rows, err := stmt.Query(startYearInt, endYearInt)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(
				&row.ChristeningsDesc,
				&row.TotalCount,
				&row.WeekNo,
				&row.WeekID,
				&row.Year)
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
		fmt.Fprintf(w, string(response))
	}

}
