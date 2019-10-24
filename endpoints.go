package relecapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Endpoint describes the endpoints available in this API.
type Endpoint struct {
	Endpoint    string `json:"endpoint"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Citation    string `json:"citation"`
}

// EndpointHandler describes the list of endpoints.
func (s *Server) EndpointHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]Endpoint, 0)
		var row Endpoint

		query := `SELECT endpoint, title, description, citation
							FROM sources 
							WHERE endpoint IS NOT NULL 
							ORDER BY endpoint;`

		rows, err := s.Database.Query(query)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Endpoint, &row.Title, &row.Description, &row.Citation)
			if err != nil {
				log.Println(err)
			}
			// Turn the endpoint slug into a proper URL
			row.Endpoint = "http://" + r.Host + "/" + row.Endpoint + "/"
			results = append(results, row)
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
		}

		response, _ := json.MarshalIndent(results, "", "  ")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(response))
	}
}
