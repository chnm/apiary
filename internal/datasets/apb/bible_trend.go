package apb

import (
	"net/http"
)

// APBBibleTrendHandler returns the rates of quotation per year for a verse.
func (h *Handler) APBBibleTrendHandler() http.HandlerFunc {

	query := `
	SELECT
		year,
		n,
		SUM(n) OVER (ORDER BY year ROWS BETWEEN 2 PRECEDING AND 2 FOLLOWING) / SUM(wordcount) OVER (ORDER BY year ROWS BETWEEN 2 PRECEDING AND 2 FOLLOWING) * 1000000 AS q_rate_smoothed
  FROM
	(SELECT series.year,
		COALESCE(n, 0) as n,
		wordcount
	FROM
	(SELECT generate_series($2::int, $3::int) AS year) series
	LEFT JOIN 
	(SELECT 
		year, 
		n, 
		wordcount
		FROM apb.rate_quotations_bible
		WHERE corpus = $1) AS q
	ON series.year = q.year 
	ORDER BY series.year) res
	`

	return func(w http.ResponseWriter, r *http.Request) {

		corpus := "chronam"
		minYear, maxYear := 1836, 1922

		results := make([]VerseTrend, 0, 87) // Preallocate slice capacity
		var row VerseTrend

		rows, err := h.db.Query(r.Context(), query, corpus, minYear, maxYear)
		if err != nil {
			internalServerError(w, "error querying Bible trend", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			if err := rows.Scan(&row.Year, &row.N, &row.QuotationRateSmooth); err != nil {
				internalServerError(w, "error scanning Bible trend", err)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			internalServerError(w, "error iterating Bible trend", err)
			return
		}

		wrapper := VerseTrendResponse{Reference: "bible", Corpus: corpus, Trend: results}
		writeJSONResponse(w, wrapper)
	}

}
