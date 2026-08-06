package apb

import (
	"net/http"
)

// VerseQuotation is a single instance of a quotation
type VerseQuotation struct {
	Reference   string  `json:"reference"`
	DocID       string  `json:"docID"`
	Date        string  `json:"date"`
	Probability float32 `json:"probability"`
	Title       string  `json:"title"`
	State       string  `json:"state"`
}

// APBVerseQuotationsHandler returns the instances of quotations for a verse.
func (h *Handler) APBVerseQuotationsHandler() http.HandlerFunc {

	query := `
	SELECT q.reference_id, q.doc_id, q.date::text, q.probability,
	 	n.title_clean, places.state
	FROM apb.quotations q
	LEFT JOIN apb.chronam_pages p ON q.doc_id = p.doc_id
	LEFT JOIN apb.chronam_newspapers n ON p.lccn = n.lccn
	LEFT JOIN (SELECT DISTINCT ON (lccn) lccn, state FROM apb.chronam_newspaper_places ORDER BY lccn) places ON p.lccn = places.lccn
	WHERE reference_id = $1 AND corpus = 'chronam'
	ORDER BY date;
	`

	return func(w http.ResponseWriter, r *http.Request) {

		refs := r.URL.Query()["ref"]

		results := make([]VerseQuotation, 0)
		var row VerseQuotation

		rows, err := h.db.Query(r.Context(), query, refs[0])
		if err != nil {
			internalServerError(w, "error querying verse quotations", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			if err := rows.Scan(&row.Reference, &row.DocID, &row.Date, &row.Probability, &row.Title, &row.State); err != nil {
				internalServerError(w, "error scanning verse quotation", err)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			internalServerError(w, "error iterating verse quotations", err)
			return
		}

		if len(results) == 0 {
			http.Error(w, "404 Not found.", http.StatusNotFound)
			return
		}

		writeJSONResponse(w, results)
	}

}
