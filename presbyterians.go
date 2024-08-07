package apiary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// PresbyteriansByYear holds aggregate data on Presbyterian membership and churches.
type PresbyteriansByYear struct {
	Year     int `json:"year"`
	Members  int `json:"members"`
	Churches int `json:"churches"`
}

// PresbyteriansHandler returns the aggregate data on Presbyterian memberhsip and churches.
func (s *Server) PresbyteriansHandler() http.HandlerFunc {

	query := `
	SELECT 
		year, 
		SUM(members) as members, 
		SUM(churches) as churches
	FROM presbyterians_weber 
	WHERE members IS NOT NULL 
	GROUP BY year 
	ORDER BY year;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]PresbyteriansByYear, 0)
		var row PresbyteriansByYear

		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Year, &row.Members, &row.Churches)
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
