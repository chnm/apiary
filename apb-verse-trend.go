package dataapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// VerseTrend is the rate of quotations in a single year for a single verse in a given corpus.
type VerseTrend struct {
	Year              int     `json:"year"`
	Corpus            string  `json:"corpus"`
	N                 int     `json:"n"`
	QuotationsPerPage float64 `json:"q_per_page_e3"`
	QuotationsPerWord float64 `json:"q_per_word_e6"`
}

// VerseTrendHandler returns the rates of quotation per year for a verse
func (s *Server) VerseTrendHandler() http.HandlerFunc {

	query := `
	SELECT series.year, series.corpus,
		COALESCE(n, 0) as N,
		COALESCE(q_per_page_e3, 0) AS q_per_page_e3,
		COALESCE(q_per_word_e6, 0) AS q_per_page_e6
	FROM
	(SELECT generate_series(1789, 1963) AS year, 'chronam'::text AS corpus
		UNION ALL
		SELECT generate_series(1800, 1899) AS year, 'ncnp'::text AS corpus) AS series
	LEFT JOIN 
	(SELECT year, corpus, n, q_per_page_e3, q_per_word_e6 
		FROM apb.rate_quotations_verses 
		WHERE reference_id = $1) AS q
	ON series.year = q.year AND series.corpus = q.corpus
	ORDER BY series.corpus, series.year
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
			err := rows.Scan(&row.Year, &row.Corpus, &row.N, &row.QuotationsPerPage, &row.QuotationsPerWord)
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
