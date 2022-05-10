package apiary

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// BibleBook describes a book of the Bible and which part of the Bible it is in.
type BibleBook struct {
	Book  string `json:"book"`
	Part  string `json:"part"`
	Order int    `json:"order"`
}

// APBBibleBooksHandler returns the books of the Bible (in the KJV).
func (s *Server) APBBibleBooksHandler() http.HandlerFunc {

	query := `
	SELECT DISTINCT book, part, book_order
	FROM apb.scriptures
	WHERE version = 'KJV'
	ORDER BY book_order;
	`

	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["bible-books"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {

		var result []BibleBook
		var book BibleBook

		rows, err := stmt.Query()
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&book.Book, &book.Part, &book.Order)
			if err != nil {
				log.Println(err)
			}
			result = append(result, book)
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
		}

		response, _ := json.Marshal(result)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(response))
	}

}
