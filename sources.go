package dataapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Source describes the sources available in this API and provides a sample URL.
type Source struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Citation    string `json:"citation"`
	APIURL      string `json:"api_url"`
}

// SourcesHandler describes the sources that are available in this API, with
// sample URLs to see how the API works.
func (s *Server) SourcesHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]Source, 0)
		var row Source

		query := `SELECT title, description, citation, api_url
							FROM sources 
							WHERE api_url IS NOT NULL 
							ORDER BY api_url;`

		rows, err := s.Database.Query(query)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Title, &row.Description, &row.Citation, &row.APIURL)
			if err != nil {
				log.Println(err)
			}
			// Turn the endpoint slug into a proper URL
			row.APIURL = "http://" + r.Host + row.APIURL
			results = append(results, row)
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, _ := json.MarshalIndent(results, "", "  ")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(response))
	}
}
