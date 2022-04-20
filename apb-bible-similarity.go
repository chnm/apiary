package dataapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// BibleSimilarityEdge describes an edge between two parts of the Bible
type BibleSimilarityEdge struct {
	A string `json:"source"`
	B string `json:"target"`
	N int    `json:"n"`
}

// BibleSimilarityHandler returns the information about the network of
// similarities within the Bible.
func (s *Server) BibleSimilarityHandler() http.HandlerFunc {

	edgeQuery := `
	SELECT 
	a_book AS a,
	b_book AS b,
	COUNT(*) AS n
	FROM 
	(
	SELECT
	s1.book AS a_book,
	s2.book AS b_book
	FROM apb.scriptures_intraversion_pairs p
	LEFT JOIN
	apb.scriptures s1
	ON p.a = s1.verse_id
	LEFT JOIN
	apb.scriptures s2
	ON p.b = s2.verse_id
	WHERE 
	p.version = 'KJV' AND p.score > 0.18
	) AS pairs
	GROUP BY a_book, b_book
	HAVING COUNT(*) >= 5 AND a_book != b_book
	`
	edgeStmt, err := s.Database.Prepare(edgeQuery)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["apb-edges"] = edgeStmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {

		var edge BibleSimilarityEdge
		var result []BibleSimilarityEdge

		rows, err := edgeStmt.Query()
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&edge.A, &edge.B, &edge.N)
			if err != nil {
				log.Println(err)
			}
			result = append(result, edge)
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
