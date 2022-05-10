package apiary

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// This file creates a series of endpoints to return all possible names for
// populated places, associated with their county IDs from AHCB as of 1926-1928.

// PlaceCounty represents a county with ID and name
type PlaceCounty struct {
	CountyAHCB string `json:"county_ahcb"`
	County     string `json:"name"`
}

// Place represents a place with ID and name
type Place struct {
	PlaceID int    `json:"place_id"`
	Place   string `json:"place"`
	MapName string `json:"map_name"`
}

// PlaceDetails represents details about a place
type PlaceDetails struct {
	PlaceID    int    `json:"place_id"`
	Place      string `json:"place"`
	MapName    string `json:"map_name"`
	County     string `json:"county"`
	CountyAHCB string `json:"county_ahcb"`
	State      string `json:"state"`
}

// CountiesInState returns a list of all the counties in a state, with
// IDs from AHCB.
func (s *Server) CountiesInState() http.HandlerFunc {

	query := `
		SELECT DISTINCT county_ahcb, county
		FROM popplaces_1926
		WHERE state = $1
		ORDER BY county;
		`

	return func(w http.ResponseWriter, r *http.Request) {

		state := mux.Vars(r)["state"]
		state = strings.ToUpper(state)

		results := make([]PlaceCounty, 0)
		var row PlaceCounty

		rows, err := s.DB.Query(context.TODO(), query, state)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.CountyAHCB, &row.County)
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

// PlacesInCounty returns a list of all the populated places in a county.
func (s *Server) PlacesInCounty() http.HandlerFunc {

	query := `
		SELECT place_id, place, map_name
		FROM popplaces_1926
		WHERE county_ahcb = $1
		ORDER BY place;
		`

	return func(w http.ResponseWriter, r *http.Request) {

		county := mux.Vars(r)["county"]
		county = strings.ToLower(county)

		results := make([]Place, 0)
		var row Place

		rows, err := s.DB.Query(context.TODO(), query, county)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.PlaceID, &row.Place, &row.MapName)
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

// Place returns the details about a populated place.
func (s *Server) Place() http.HandlerFunc {

	query := `
		SELECT place_id, place, map_name, county, county_ahcb, state
		FROM popplaces_1926
		WHERE place_id = $1
		`

	return func(w http.ResponseWriter, r *http.Request) {

		placeID, err := strconv.Atoi(mux.Vars(r)["place"])
		if err != nil {
			http.Error(w, "Bad request: place ID must be an integer", http.StatusBadRequest)
			return
		}

		var result PlaceDetails

		err = s.DB.QueryRow(context.TODO(), query, placeID).Scan(&result.PlaceID, &result.Place,
			&result.MapName, &result.County, &result.CountyAHCB, &result.State)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, fmt.Sprintf("Not found: No place with id %v.", placeID), http.StatusNotFound)
				return
			}
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		response, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(response))

	}
}
