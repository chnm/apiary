package apiary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// ParishByYear describes a parish's names and various metadata for a given parish.
type ParishByYear struct {
	ParishName string     `json:"name"`
	CountType  string     `json:"count_type"`
	TotalCount NullInt64  `json:"count"`
	StartDay   NullInt64  `json:"start_day"`
	StartMonth NullString `json:"start_month"`
	EndDay     int        `json:"end_day"`
	EndMonth   string     `json:"end_month"`
	Year       int        `json:"year"`
	WeekNo     int        `json:"week_no"`
	WeekID     string     `json:"week_id"`
}

// BillsHandler returns ...
func (s *Server) BillsHandler() http.HandlerFunc {

	query := `
	SELECT
		p.canonical_name,
		b.count_type,
		b.count,
		w.start_day,
		w.start_month,
		w.end_day,
		w.end_month,
		y.year,
		w.week_no,
		b.week_id
	FROM
		bom.bill_of_mortality b
	JOIN
		bom.parishes p ON p.id = b.parish_id
	JOIN
		bom.year y ON y.year_id = b.year_id
	JOIN
		bom.week w ON w.week_id = b.week_id
	WHERE
		year >= $1
		AND year <= $2
		AND b.bill_type = 'Weekly'
	ORDER BY
		canonical_name;
	`

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

		results := make([]ParishByYear, 0)
		var row ParishByYear

		rows, err := s.DB.Query(context.TODO(), query, startYearInt, endYearInt)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(
				&row.ParishName,
				&row.CountType,
				&row.TotalCount,
				&row.StartDay,
				&row.StartMonth,
				&row.EndDay,
				&row.EndMonth,
				&row.Year,
				&row.WeekNo,
				&row.WeekID)
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
