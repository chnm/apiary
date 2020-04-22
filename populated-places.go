package dataapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// This file creates a series of endpoints to return all possible names for
// populated places, associated with their county IDs from AHCB as of 1926-1928.

// PopPlacesCounty represents a county in a state, with its AHCB ID.
type PopPlacesCounty struct {
	Name       string `json:"name"`
	CountyAHCB string `json:"county_ahcb"`
}

// PopPlacesCountiesInState returns a list of all the counties in a state, with
// IDs from AHCB.
func (s *Server) PopPlacesCountiesInState() http.HandlerFunc {

	query := `
		SELECT DISTINCT county, county_ahcb
		FROM popplaces_1926
		WHERE state = $1
		ORDER BY county;
		`

	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["pop-places-counties-in-state"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {

		state := mux.Vars(r)["state"]
		state = strings.ToUpper(state)

		results := make([]PopPlacesCounty, 0)
		var row PopPlacesCounty

		rows, err := stmt.Query(state)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Name, &row.CountyAHCB)
			if err != nil {
				log.Println(err)
			}
			results = append(results, row)

		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		response, _ := json.Marshal(results)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(response))

	}
}
