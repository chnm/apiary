package apiary

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Verse describes the reference and text of a single Bible verse
type Verse struct {
	Reference string   `json:"reference"`
	Text      string   `json:"text"`
	Related   []string `json:"related"`
}

// APBVerseHandler returns information about a verse, and other verses which are related to it, if any.
func (s *Server) APBVerseHandler() http.HandlerFunc {

	// This query currently returns all verses which have not been
	// explicitly disallowed, rather than returning only verses which
	// have been explicitly allowed.
	verseQuery := `
	SELECT v.reference_id, s.text
	FROM apb.verse_cleanup v
	LEFT JOIN apb.scriptures s
		ON v.reference_id=s.reference_id
	WHERE 
		v.reference_id = $1 AND
		(v.use = TRUE OR v.use IS NULL) AND
		s.version = 'KJV';
	`

	relatedVerseQuery := `
	SELECT reference_id
	FROM apb.verse_cleanup
	WHERE reference_use = $1 AND reference_id != reference_use
	`

	return func(w http.ResponseWriter, r *http.Request) {

		refs := r.URL.Query()["ref"]

		var result Verse

		err := s.DB.QueryRow(context.TODO(), verseQuery, refs[0]).Scan(&result.Reference, &result.Text)
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 Not found."))
			return
		} else if err != nil {
			log.Println(err)
		}

		related := make([]string, 0)
		rows, err := s.DB.Query(context.TODO(), relatedVerseQuery, refs[0])
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		var rel string
		for rows.Next() {
			err := rows.Scan(&rel)
			if err != nil {
				log.Println(err)
			}
			related = append(related, rel)
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
		}

		result.Related = related

		response, _ := json.Marshal(result)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(response))
	}

}
