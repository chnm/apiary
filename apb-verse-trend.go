package dataapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// VerseTrend is the rate of quotations in a single year for a single verse in a given corpus. The quotation rate is expressed in quotations per million words; the smoothed rate has the same units, and is a centered three-year rolling average.
type VerseTrend struct {
	Year                int     `json:"year"`
	N                   int     `json:"n"`
	QuotationRateSmooth float64 `json:"smoothed"`
}

// VerseTrendResponse wraps the data with the queries that were made
type VerseTrendResponse struct {
	Reference string       `json:"reference"`
	Corpus    string       `json:"corpus"`
	Trend     []VerseTrend `json:"trend"`
}

// VerseTrendHandler returns the rates of quotation per year for a verse
func (s *Server) VerseTrendHandler() http.HandlerFunc {

	query := `
	SELECT 
	res.year,
	n,
	SUM(n) OVER (ORDER BY res.year ROWS BETWEEN 2 PRECEDING AND 2 FOLLOWING) / SUM(wordcount) OVER (ORDER BY res.year ROWS BETWEEN 2 PRECEDING AND 2 FOLLOWING) * 1000000 AS q_rate_smoothed
	FROM
		(SELECT series.year,
			COALESCE(n, 0) as n
			FROM
				(SELECT generate_series($3::int, $4::int) AS year) series
				LEFT JOIN 
				(SELECT year, n, corpus, reference_id
				FROM apb.count_quotations_verses 
				WHERE corpus = $1 AND reference_id = $2) AS q
				ON series.year = q.year 
				ORDER BY series.year) res
				LEFT JOIN 
			(SELECT year, wordcount FROM apb.wordcounts
			WHERE corpus = 'chronam') wc
			ON res.year = wc.year
	`

	stmt, err := s.APB.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["apb-verse-trend"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {

		// Return a 404 error if we don't get exactly one reference
		queryRef := r.URL.Query()["ref"]
		var ref string
		if len(queryRef) == 1 {
			ref = queryRef[0]
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("400 Bad request. Please provide exactly one reference."))
			return
		}

		// Use chronam as the default corpus
		corpusRef := r.URL.Query()["corpus"]
		corpus := "chronam"
		if len(corpusRef) > 0 {
			corpus = corpusRef[0]
			if !(corpus == "ncnp" || corpus == "chronam") {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("400 Bad request. Corpus must be 'ncnp' or 'chronam'."))
				return
			}
		}

		var minYear, maxYear int
		if corpus == "chronam" {
			// Years in which there is a major disjuncture in the total number of pages
			minYear = 1836
			maxYear = 1922
		} else if corpus == "ncnp" {
			minYear = 1836
			maxYear = 1899
		}

		results := make([]VerseTrend, 0, 175) // Preallocate slice capacity
		var row VerseTrend

		rows, err := stmt.Query(corpus, ref, minYear, maxYear)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Year, &row.N, &row.QuotationRateSmooth)
			if err != nil {
				log.Println(err)
			}
			results = append(results, row)
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
		}

		wrapper := VerseTrendResponse{Reference: ref, Corpus: corpus, Trend: results}

		response, _ := json.Marshal(wrapper)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(response))
	}

}
