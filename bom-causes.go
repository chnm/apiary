package apiary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// TODO: Describe
type DeathCauses struct {
	DeathID    int        `json:"death_id"`
	Death      string     `json:"death"`
	Count      NullInt64  `json:"count"`
	WeekID     string     `json:"week_id"`
	StartDay   NullString `json:"start_day"`
	StartMonth NullString `json:"start_month"`
	EndDay     NullString `json:"end_day"`
	EndMonth   NullString `json:"end_month"`
	Year       string     `json:"year"`
}

// TODO: Describe
func (s *Server) DeathCausesHandler() http.HandlerFunc {

	query := `
	SELECT 
		c.id, c.death, c.count, c.week_id, w.start_day, w.start_month, w.end_day, w.end_month, w.year_id
	FROM 
		bom.causes_of_death c
	JOIN 
		bom.week w ON w.week_id = c.week_id
	ORDER BY 
		id;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]DeathCauses, 0)
		var row DeathCauses

		rows, err := s.Pool.Query(context.TODO(), query)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.DeathID, &row.Death, &row.Count, &row.WeekID, &row.StartDay, &row.EndDay, &row.StartMonth, &row.EndMonth, &row.Year)
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
