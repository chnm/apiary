package dataapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// VerseTrend is the rate of quotations in a single year for a single verse in a given corpus.
type VerseTrend struct {
	Reference         string  `json:"reference"`
	Corpus            string  `json:"corpus"`
	Year              int     `json:"year"`
	N                 int     `json:"n"`
	QuotationsPerPage float64 `json:"q_per_page_e3"`
	QuotationsPerWord float64 `json:"q_per_word_e6"`
}

// VerseTrendHandler returns the rates of quotation per year for a verse
func (s *Server) VerseTrendHandler() http.HandlerFunc {

	query := `
	SELECT reference_id, corpus, year, 
				 n, q_per_page_e3, q_per_word_e6 
	FROM apb.rate_quotations_verses 
	WHERE reference_id = $1;
	`
	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["apb-verse-trend"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {

		refs := r.URL.Query()["ref"]

		results := make([]VerseTrend, 0, 128) // Preallocate slice capacity
		var row VerseTrend

		rows, err := stmt.Query(refs[0])
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Reference, &row.Corpus, &row.Year, &row.N, &row.QuotationsPerPage, &row.QuotationsPerWord)
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
