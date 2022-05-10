package apiary

import (
	"encoding/json"
	"fmt"
	"log"
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

// APBIndexFeaturedHandler returns featured verses.
func (s *Server) APBIndexFeaturedHandler() http.HandlerFunc {

	query := `
	SELECT t.reference_id, s.text, t.n
	FROM apb.top_verses t
	LEFT JOIN apb.scriptures s ON t.reference_id = s.reference_id
	LEFT JOIN apb.verse_cleanup c ON t.reference_id = c.reference_use
	WHERE s.version = 'KJV' AND c.display = True
  ORDER BY s.book_order, s.chapter, s.verse;
	`

	stmt, err := s.Database.Prepare(query)

	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["apb-index-featured"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {
		var results []APBIndexItem
		var row APBIndexItem

		rows, err := stmt.Query()
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Reference, &row.Text, &row.Count)
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

// APBIndexBiblicalOrderHandler returns verses in their biblical order.
func (s *Server) APBIndexBiblicalOrderHandler() http.HandlerFunc {

	query := `
	SELECT t.reference_id, s.text, t.n
	FROM apb.top_verses t
	LEFT JOIN apb.verse_cleanup c ON t.reference_id = c.reference_id
	LEFT JOIN apb.scriptures s ON t.reference_id = s.reference_id
	WHERE t.n > 500 AND c.use = TRUE AND s.version = 'KJV' AND s.part != 'Apocrypha'
  ORDER BY s.book_order, s.chapter, s.verse;
	`

	stmt, err := s.Database.Prepare(query)

	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["apb-index-biblical-order"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {
		var results []APBIndexItem
		var row APBIndexItem

		rows, err := stmt.Query()
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Reference, &row.Text, &row.Count)
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

// APBIndexTopHandler returns top verses.
func (s *Server) APBIndexTopHandler() http.HandlerFunc {

	query := `
	SELECT t.reference_id, s.text, t.n
	FROM apb.top_verses t
	LEFT JOIN apb.scriptures s ON t.reference_id = s.reference_id
	WHERE s.version = 'KJV' AND s.part != 'Apocrypha'
	ORDER BY t.n DESC
	LIMIT 100;
	`

	stmt, err := s.Database.Prepare(query)

	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["apb-index-top"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {
		var results []APBIndexItem
		var row APBIndexItem

		rows, err := stmt.Query()
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Reference, &row.Text, &row.Count)
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

// APBIndexChronologicalHandler returns verses in chronological order by their peak.
func (s *Server) APBIndexChronologicalHandler() http.HandlerFunc {

	query := `
	SELECT t.reference_id, s.text, t.n, p.year
	FROM apb.top_verses t
	LEFT JOIN apb.verse_cleanup c ON t.reference_id = c.reference_id
	LEFT JOIN apb.scriptures s ON t.reference_id = s.reference_id
  LEFT JOIN apb.verse_peaks p ON t.reference_id = p.reference_id
	WHERE t.n > 500 AND c.use = TRUE AND s.version = 'KJV' AND s.part != 'Apocrypha'
  ORDER BY p.year, t.n DESC;
	`

	stmt, err := s.Database.Prepare(query)

	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["apb-index-chronological-order"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {
		var results []APBIndexItemWithYear
		var row APBIndexItemWithYear

		rows, err := stmt.Query()
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Reference, &row.Text, &row.Count, &row.Peak)
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

// APBIndexAllHandler returns basically all available verses in their biblical order.
func (s *Server) APBIndexAllHandler() http.HandlerFunc {

	query := `
	SELECT t.reference_id, s.text
	FROM apb.top_verses t
	LEFT JOIN apb.verse_cleanup c ON t.reference_id = c.reference_id
	LEFT JOIN apb.scriptures s ON t.reference_id = s.reference_id
	WHERE t.n > 100 AND c.use = TRUE AND s.version = 'KJV' AND s.part != 'Apocrypha'
  ORDER BY s.book_order, s.chapter, s.verse;
	`

	stmt, err := s.Database.Prepare(query)

	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["apb-index-all"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {
		var results []APBIndexItemText
		var row APBIndexItemText

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
