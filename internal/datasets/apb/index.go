package apb

import (
	"net/http"
)

// APBIndexItem is an entry in one of the different indexes to verses
type APBIndexItem struct {
	Reference string `json:"reference"`
	Text      string `json:"text"`
	Count     int    `json:"count"`
}

// APBIndexItemText is an entry in one of the different indexes to verses, with
// the reference and the text of the verse.
type APBIndexItemText struct {
	Reference string `json:"reference"`
	Text      string `json:"text"`
}

// APBIndexItemWithYear is an index item with the peak year
type APBIndexItemWithYear struct {
	Reference string `json:"reference"`
	Text      string `json:"text"`
	Count     int    `json:"count"`
	Peak      int    `json:"peak"`
}

// APBIndexFeaturedHandler returns featured verses for APB.
func (h *Handler) APBIndexFeaturedHandler() http.HandlerFunc {

	query := `
	SELECT t.reference_id, s.text, t.n
	FROM apb.top_verses t
	LEFT JOIN apb.scriptures s ON t.reference_id = s.reference_id
	LEFT JOIN apb.verse_cleanup c ON t.reference_id = c.reference_use
	WHERE s.version = 'KJV' AND c.display = True
  ORDER BY s.book_order, s.chapter, s.verse;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		var results []APBIndexItem
		var row APBIndexItem

		rows, err := h.db.Query(r.Context(), query)
		if err != nil {
			internalServerError(w, "error querying featured verse index", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			if err := rows.Scan(&row.Reference, &row.Text, &row.Count); err != nil {
				internalServerError(w, "error scanning featured verse index", err)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			internalServerError(w, "error iterating featured verse index", err)
			return
		}

		writeJSONResponse(w, results)
	}

}

// APBIndexBiblicalOrderHandler returns verses in their biblical order.
func (h *Handler) APBIndexBiblicalOrderHandler() http.HandlerFunc {

	query := `
	SELECT t.reference_id, s.text, t.n
	FROM apb.top_verses t
	LEFT JOIN apb.verse_cleanup c ON t.reference_id = c.reference_id
	LEFT JOIN apb.scriptures s ON t.reference_id = s.reference_id
	WHERE t.n > 500 AND c.use = TRUE AND s.version = 'KJV' AND s.part != 'Apocrypha'
  ORDER BY s.book_order, s.chapter, s.verse;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		var results []APBIndexItem
		var row APBIndexItem

		rows, err := h.db.Query(r.Context(), query)
		if err != nil {
			internalServerError(w, "error querying biblical verse index", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			if err := rows.Scan(&row.Reference, &row.Text, &row.Count); err != nil {
				internalServerError(w, "error scanning biblical verse index", err)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			internalServerError(w, "error iterating biblical verse index", err)
			return
		}

		writeJSONResponse(w, results)
	}

}

// APBIndexTopHandler returns top verses for APB.
func (h *Handler) APBIndexTopHandler() http.HandlerFunc {

	query := `
	SELECT t.reference_id, s.text, t.n
	FROM apb.top_verses t
	LEFT JOIN apb.scriptures s ON t.reference_id = s.reference_id
	WHERE s.version = 'KJV' AND s.part != 'Apocrypha'
	ORDER BY t.n DESC
	LIMIT 100;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		var results []APBIndexItem
		var row APBIndexItem

		rows, err := h.db.Query(r.Context(), query)
		if err != nil {
			internalServerError(w, "error querying top verse index", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			if err := rows.Scan(&row.Reference, &row.Text, &row.Count); err != nil {
				internalServerError(w, "error scanning top verse index", err)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			internalServerError(w, "error iterating top verse index", err)
			return
		}

		writeJSONResponse(w, results)
	}

}

// APBIndexChronologicalHandler returns verses in chronological order by their peak.
func (h *Handler) APBIndexChronologicalHandler() http.HandlerFunc {

	query := `
	SELECT t.reference_id, s.text, t.n, p.year
	FROM apb.top_verses t
	LEFT JOIN apb.verse_cleanup c ON t.reference_id = c.reference_id
	LEFT JOIN apb.scriptures s ON t.reference_id = s.reference_id
  LEFT JOIN apb.verse_peaks p ON t.reference_id = p.reference_id
	WHERE t.n > 500 AND c.use = TRUE AND s.version = 'KJV' AND s.part != 'Apocrypha'
  ORDER BY p.year, t.n DESC;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		var results []APBIndexItemWithYear
		var row APBIndexItemWithYear

		rows, err := h.db.Query(r.Context(), query)
		if err != nil {
			internalServerError(w, "error querying chronological verse index", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			if err := rows.Scan(&row.Reference, &row.Text, &row.Count, &row.Peak); err != nil {
				internalServerError(w, "error scanning chronological verse index", err)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			internalServerError(w, "error iterating chronological verse index", err)
			return
		}

		writeJSONResponse(w, results)
	}

}

// APBIndexAllHandler returns basically all available verses in their biblical order.
func (h *Handler) APBIndexAllHandler() http.HandlerFunc {

	query := `
	SELECT t.reference_id, s.text
	FROM apb.top_verses t
	LEFT JOIN apb.verse_cleanup c ON t.reference_id = c.reference_id
	LEFT JOIN apb.scriptures s ON t.reference_id = s.reference_id
	WHERE t.n > 100 AND c.use = TRUE AND s.version = 'KJV' AND s.part != 'Apocrypha'
  ORDER BY s.book_order, s.chapter, s.verse;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		var results []APBIndexItemText
		var row APBIndexItemText

		rows, err := h.db.Query(r.Context(), query)
		if err != nil {
			internalServerError(w, "error querying complete verse index", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			if err := rows.Scan(&row.Reference, &row.Text); err != nil {
				internalServerError(w, "error scanning complete verse index", err)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			internalServerError(w, "error iterating complete verse index", err)
			return
		}

		writeJSONResponse(w, results)
	}

}
