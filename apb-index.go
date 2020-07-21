package dataapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// IndexItem is an entry in one of the different indexes to verses
type IndexItem struct {
	Reference string `json:"reference"`
	Text      string `json:"text"`
}

// APBIndexFeaturedHandler returns featured verses.
func (s *Server) APBIndexFeaturedHandler() http.HandlerFunc {

	query := `
	SELECT v.reference_use, s.text
	FROM apb.verse_cleanup v
	LEFT JOIN apb.scriptures s
	ON v.reference_use = s.reference_id
	WHERE v.display = True AND s.version = 'KJV'
	ORDER BY s.book_order, s.chapter, s.verse;
	`

	stmt, err := s.Database.Prepare(query)

	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["apb-index-featured"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {
		var results []IndexItem
		var row IndexItem

		rows, err := stmt.Query()
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Reference, &row.Text)
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
