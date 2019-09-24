package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// PresbyteriansHandler returns the number of Presbyerians
func (s *Server) PresbyteriansHandler() http.HandlerFunc {

	type presbyteriansByYear struct {
		Year    int `json:"year"`
		Members int `json:"members"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]presbyteriansByYear, 0)
		var row presbyteriansByYear

		rows, err := s.Database.Query("SELECT year, SUM(members) as members FROM presbyterians_weber WHERE members IS NOT NULL GROUP BY year ORDER BY year;")
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Year, &row.Members)
			if err != nil {
				log.Println(err)
			}
			results = append(results, row)

		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
		}

		response, _ := json.Marshal(results)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(response))
	}
}
