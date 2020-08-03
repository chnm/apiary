package dataapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// BibleTrendHandler returns the rates of quotation per year for a verse
func (s *Server) BibleTrendHandler() http.HandlerFunc {

	query := `
	SELECT
		year,
		n,
		q_per_word_e6,
		AVG(q_per_word_e6) OVER (ORDER BY year ROWS BETWEEN 2 PRECEDING AND 2 FOLLOWING) AS q_rate_smoothed
  FROM
	(SELECT series.year,
		COALESCE(n, 0) as n,
		COALESCE(q_per_word_e6, 0) AS q_per_word_e6
	FROM
	(SELECT generate_series($2::int, $3::int) AS year) series
	LEFT JOIN 
	(SELECT year, n, q_per_word_e6 
		FROM apb.rate_quotations_bible2
		WHERE corpus = $1) AS q
	ON series.year = q.year 
	ORDER BY series.year) res
	`

	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["apb-bible-trend"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {

		corpus := "chronam"
		minYear, maxYear := 1836, 1922

		results := make([]VerseTrend, 0, 87) // Preallocate slice capacity
		var row VerseTrend

		rows, err := stmt.Query(corpus, minYear, maxYear)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Year, &row.N, &row.QuotationRate, &row.QuotationRateSmooth)
			if err != nil {
				log.Println(err)
			}
			results = append(results, row)
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
		}

		wrapper := VerseTrendResponse{Reference: "bible", Corpus: corpus, Trend: results}

		response, _ := json.Marshal(wrapper)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(response))
	}

}
