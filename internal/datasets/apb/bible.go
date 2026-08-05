package apb

import (
	"net/http"
)

// BibleBook describes a book of the Bible and which part of the Bible it is in.
type BibleBook struct {
	Book  string `json:"book"`
	Part  string `json:"part"`
	Order int    `json:"order"`
}

// APBBibleBooksHandler returns the books of the Bible (in the KJV).
func (h *Handler) APBBibleBooksHandler() http.HandlerFunc {

	query := `
	SELECT DISTINCT book, part, book_order
	FROM apb.scriptures
	WHERE version = 'KJV'
	ORDER BY book_order;
	`

	return func(w http.ResponseWriter, r *http.Request) {

		var result []BibleBook
		var book BibleBook

		rows, err := h.db.Query(r.Context(), query)
		if err != nil {
			internalServerError(w, "error querying Bible books", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			if err := rows.Scan(&book.Book, &book.Part, &book.Order); err != nil {
				internalServerError(w, "error scanning Bible book", err)
				return
			}
			result = append(result, book)
		}
		if err := rows.Err(); err != nil {
			internalServerError(w, "error iterating Bible books", err)
			return
		}

		writeJSONResponse(w, result)
	}

}
