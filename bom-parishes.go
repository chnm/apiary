package apiary

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Parish describes a denomination's names and various systems of classification.
type Parish struct {
	ParishID      int    `json:"id"`
	Name          string `json:"name"`
	CanonicalName string `json:"canonical_name"`
}

// ParishesHandler returns
func (s *Server) ParishesHandler() http.HandlerFunc {

	query := `
	SELECT id, parish_name, canonical_name 
	FROM bom.parishes
	ORDER BY canonical_name;
	`

	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["parishes"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]Parish, 0)
		var row Parish

		rows, err := stmt.Query()
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.ParishID, &row.Name, &row.CanonicalName)
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
