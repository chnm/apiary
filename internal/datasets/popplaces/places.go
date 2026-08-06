package popplaces

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

// This file creates a series of endpoints to return all possible names for
// populated places, associated with their county IDs from AHCB as of 1926-1928.

// PlaceCounty represents a county with ID and name
type PlaceCounty struct {
	CountyAHCB string `json:"county_ahcb"`
	County     string `json:"name"`
}

// Place represents a populated place with its ID, name, and coordinates.
type Place struct {
	PlaceID int     `json:"place_id"`
	Place   string  `json:"place"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

// PlaceDetails represents details about a populated place.
type PlaceDetails struct {
	PlaceID    int     `json:"place_id"`
	Place      string  `json:"place"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	County     string  `json:"county"`
	CountyAHCB string  `json:"county_ahcb"`
	State      string  `json:"state"`
}

// CountiesInState returns a list of all the counties in a state, with
// IDs from AHCB.
func (h *Handler) CountiesInState() http.HandlerFunc {

	query := `
		SELECT DISTINCT county_ahcb, county
		FROM relcensus.popplaces_1926
		WHERE state = $1
		ORDER BY county;
		`

	return func(w http.ResponseWriter, r *http.Request) {

		state := mux.Vars(r)["state"]
		state = strings.ToUpper(state)

		results := make([]PlaceCounty, 0)

		rows, err := h.db.Query(r.Context(), query, state)
		if err != nil {
			log.Printf("query populated-place counties: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var row PlaceCounty
			if err := rows.Scan(&row.CountyAHCB, &row.County); err != nil {
				log.Printf("scan populated-place county: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			log.Printf("iterate populated-place counties: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(results)
		if err != nil {
			log.Printf("marshal populated-place counties: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(response); err != nil {
			log.Printf("write populated-place counties response: %v", err)
		}
	}
}

// PlacesInCounty returns a list of all the populated places in a county.
func (h *Handler) PlacesInCounty() http.HandlerFunc {

	query := `
		SELECT place_id, place, lat, lon
		FROM relcensus.popplaces_1926
		WHERE county_ahcb = $1
		ORDER BY place;
		`

	return func(w http.ResponseWriter, r *http.Request) {

		county := mux.Vars(r)["county"]
		county = strings.ToLower(county)

		results := make([]Place, 0)

		rows, err := h.db.Query(r.Context(), query, county)
		if err != nil {
			log.Printf("query populated places: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var row Place
			if err := rows.Scan(&row.PlaceID, &row.Place, &row.Lat, &row.Lon); err != nil {
				log.Printf("scan populated place: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			log.Printf("iterate populated places: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(results)
		if err != nil {
			log.Printf("marshal populated places: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(response); err != nil {
			log.Printf("write populated places response: %v", err)
		}
	}
}

// Place returns the details about a populated place.
func (h *Handler) Place() http.HandlerFunc {

	query := `
		SELECT place_id, place, lat, lon, county, county_ahcb, state
		FROM relcensus.popplaces_1926
		WHERE place_id = $1
		`

	return func(w http.ResponseWriter, r *http.Request) {

		placeID, err := strconv.Atoi(mux.Vars(r)["place"])
		if err != nil {
			http.Error(w, "Bad request: place ID must be an integer", http.StatusBadRequest)
			return
		}

		var result PlaceDetails

		err = h.db.QueryRow(r.Context(), query, placeID).Scan(&result.PlaceID, &result.Place,
			&result.Lat, &result.Lon, &result.County, &result.CountyAHCB, &result.State)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, fmt.Sprintf("Not found: No place with id %v.", placeID), http.StatusNotFound)
				return
			}
			log.Printf("query populated-place details: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(result)
		if err != nil {
			log.Printf("marshal populated-place details: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(response); err != nil {
			log.Printf("write populated-place details response: %v", err)
		}
	}
}
