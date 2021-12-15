package dataapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Parish describes a parish's names and various metadata for a given parish.
type ParishByYear struct {
	ParishName string    `json:"name"`
	CountType  string    `json:"count_type"`
	TotalCount NullInt64 `json:"count"`
	Year       int       `json:"year"`
	WeekNo     int       `json:"week_no"`
	WeekID     string    `json:"week_id"`
}

// BillsHandler returns ...
func (s *Server) BillsHandler() http.HandlerFunc {

	query := `
	SELECT
		p.name,
		b.count_type,
		b.count,
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
		year = $1
	ORDER BY
		name;
	`

	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["bills-of-mortality"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {
		year := r.URL.Query().Get("year")

		if year == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		yearInt, err := strconv.Atoi(year)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		results := make([]ParishByYear, 0)
		var row ParishByYear

		rows, err := stmt.Query(yearInt)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.ParishName, &row.CountType, &row.TotalCount,
				&row.Year, &row.WeekNo, &row.WeekID)
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
