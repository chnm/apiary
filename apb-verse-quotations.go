package dataapi

import (
	"encoding/json"
	"fmt"
	"log"
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

// VerseQuotationsHandler returns the instances of quotations for a verse
func (s *Server) VerseQuotationsHandler() http.HandlerFunc {

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
	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["apb-verse-quotations"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {

		refs := r.URL.Query()["ref"]

		results := make([]VerseQuotation, 0)
		var row VerseQuotation

		rows, err := stmt.Query(refs[0])
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Reference, &row.DocID, &row.Date, &row.Probability, &row.Title, &row.State)
			if err != nil {
				log.Println(err)
			}
			results = append(results, row)
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
		}

		if len(results) == 0 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 Not found."))
		}

		response, _ := json.Marshal(results)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(response))
	}

}
