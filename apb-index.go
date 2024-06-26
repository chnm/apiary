package apiary

import (
	"context"
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

// APBIndexFeaturedHandler returns featured verses for APB.
func (s *Server) APBIndexFeaturedHandler() http.HandlerFunc {

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

		rows, err := s.DB.Query(context.TODO(), query)
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
		fmt.Fprint(w, string(response))
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

	return func(w http.ResponseWriter, r *http.Request) {
		var results []APBIndexItem
		var row APBIndexItem

		rows, err := s.DB.Query(context.TODO(), query)
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
		fmt.Fprint(w, string(response))
	}

}

// APBIndexTopHandler returns top verses for APB.
func (s *Server) APBIndexTopHandler() http.HandlerFunc {

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

		rows, err := s.DB.Query(context.TODO(), query)
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
		fmt.Fprint(w, string(response))
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

	return func(w http.ResponseWriter, r *http.Request) {
		var results []APBIndexItemWithYear
		var row APBIndexItemWithYear

		rows, err := s.DB.Query(context.TODO(), query)
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
		fmt.Fprint(w, string(response))
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

	return func(w http.ResponseWriter, r *http.Request) {
		var results []APBIndexItemText
		var row APBIndexItemText

		rows, err := s.DB.Query(context.TODO(), query)
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
		fmt.Fprint(w, string(response))
	}

}
