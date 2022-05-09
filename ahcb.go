package apiary

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

// AHCBStatesHandler returns a GeoJSON FeatureCollection containing states from
// AHCB. The handler will get the county boundaries for a particular date.
func (s *Server) AHCBStatesHandler() http.HandlerFunc {

	// The minimum and maximum dates are the range of dates for states in AHCB.
	minDate, _ := time.Parse("2006-01-02", "1783-09-03")
	maxDate, _ := time.Parse("2006-01-02", "2000-12-31")

	// Build the GeoJSON in the query itself. This query will be sent to the
	// database as a prepared statement.
	query := `
		SELECT json_build_object(
			'type','FeatureCollection',
			'features', json_agg(us_states.feature)
		)
		FROM (
			SELECT json_build_object(
				'type', 'Feature',
				'id', id,
				'geometry', ST_AsGeoJSON(geom_01)::json,
				'properties', json_build_object(
					'name', name,
					'abbr', abbr_name,
					'area_sqmi', area_sqmi,
					'terr_type', terr_type)
			) AS feature
			FROM ahcb_states
			WHERE start_date <= $1 AND end_date >= $1
			) AS us_states;
		`
	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["ahcb-states"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		date, err := dateInRange(params["date"], minDate, maxDate)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var result string // result will be a string containing GeoJSON
		err = stmt.QueryRow(date).Scan(&result)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, result)

	}
}

// AHCBCountiesHandler returns a GeoJSON FeatureCollection containing counties
// from AHCB. The handler will get the county boundaries for a particular date.
func (s *Server) AHCBCountiesHandler() http.HandlerFunc {
	minDate, _ := time.Parse("2006-01-02", "1629-03-04")
	maxDate, _ := time.Parse("2006-01-02", "2000-12-31")
	query := `
		SELECT json_build_object(
			'type','FeatureCollection',
			'features', json_agg(us_counties.feature)
		) FROM (
			SELECT json_build_object(
				'type', 'Feature',
				'id', id,
				'geometry', ST_AsGeoJSON(geom_01)::json,
				'properties', json_build_object(
					'name', name,
					'state_terr', state_terr,
					'state_terr_id', state_terr_id,
					'state_code', state_code,
					'area_sqmi', area_sqmi
				)
			) AS feature
			FROM ahcb_counties
			WHERE start_date <= $1 AND end_date >= $1
		) AS us_counties;
		`
	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["ahcb-counties"] = stmt
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		date, err := dateInRange(params["date"], minDate, maxDate)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		var result string
		err = stmt.QueryRow(date).Scan(&result)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, result)
	}
}

// AHCBCountiesByIDHandler returns a GeoJSON FeatureCollection containing counties
// from AHCB. The handler will get the county boundaries for a particular date and
// by county ID (or IDs if given a comma-separated string of values).
func (s *Server) AHCBCountiesByIDHandler() http.HandlerFunc {
	minDate, _ := time.Parse("2006-01-02", "1629-03-04")
	maxDate, _ := time.Parse("2006-01-02", "2000-12-31")
	query := `
		SELECT json_build_object(
			'type','FeatureCollection',
			'features', json_agg(us_counties.feature)
		) FROM (
			SELECT json_build_object(
				'type', 'Feature',
				'id', id,
				'geometry', ST_AsGeoJSON(geom_01)::json,
				'properties', json_build_object(
					'name', name,
					'state_terr', state_terr,
					'state_terr_id', state_terr_id,
					'state_code', state_code,
					'area_sqmi', area_sqmi
				)
			) AS feature
			FROM ahcb_counties
			WHERE start_date <= $1 AND end_date >= $1
			AND id = ANY($2)
		) AS us_counties;
		`
	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["ahcb-counties-by-id"] = stmt
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		date, err := dateInRange(params["date"], minDate, maxDate)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		var result string

		ids := pq.Array(strings.Split(params["id"], ","))
		err = stmt.QueryRow(date, ids).Scan(&result)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, result)
	}
}

// AHCBCountiesByStateTerrIDHandler returns a GeoJSON FeatureCollection containing
// counties from AHCB. The handler will get the county boundaries for a particular
// date and by state/territory ID (or IDs if given a comma-separated string of values).
func (s *Server) AHCBCountiesByStateTerrIDHandler() http.HandlerFunc {
	minDate, _ := time.Parse("2006-01-02", "1629-03-04")
	maxDate, _ := time.Parse("2006-01-02", "2000-12-31")
	query := `
		SELECT json_build_object(
			'type','FeatureCollection',
			'features', json_agg(us_counties.feature)
		) FROM (
			SELECT json_build_object(
				'type', 'Feature',
				'id', id,
				'geometry', ST_AsGeoJSON(geom_01)::json,
				'properties', json_build_object(
					'name', name,
					'state_terr', state_terr,
					'state_terr_id', state_terr_id,
					'state_code', state_code,
					'area_sqmi', area_sqmi
				)
			) AS feature
			FROM ahcb_counties
			WHERE start_date <= $1 AND end_date >= $1
			AND state_terr_id = ANY($2)
		) AS us_counties;
		`
	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["ahcb-counties-by-id"] = stmt
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		date, err := dateInRange(params["date"], minDate, maxDate)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		var result string
		stateTerrIds := pq.Array(strings.Split(params["state-terr-id"], ","))
		err = stmt.QueryRow(date, stateTerrIds).Scan(&result)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, result)
	}
}

// AHCBCountiesByStateCodeHandler returns a GeoJSON FeatureCollection containing
// counties from AHCB. The handler will get the county boundaries for a particular
// date and by state code (or state codes if given a comma-separated string of values).
func (s *Server) AHCBCountiesByStateCodeHandler() http.HandlerFunc {
	minDate, _ := time.Parse("2006-01-02", "1629-03-04")
	maxDate, _ := time.Parse("2006-01-02", "2000-12-31")
	query := `
		SELECT json_build_object(
			'type','FeatureCollection',
			'features', json_agg(us_counties.feature)
		) FROM (
			SELECT json_build_object(
				'type', 'Feature',
				'id', id,
				'geometry', ST_AsGeoJSON(geom_01)::json,
				'properties', json_build_object(
					'name', name,
					'state_terr', state_terr,
					'state_terr_id', state_terr_id,
					'state_code', state_code,
					'area_sqmi', area_sqmi
				)
			) AS feature
			FROM ahcb_counties
			WHERE start_date <= $1 AND end_date >= $1
			AND state_code = ANY($2)
		) AS us_counties;
		`
	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["ahcb-counties-by-state-code"] = stmt
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		date, err := dateInRange(params["date"], minDate, maxDate)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		var result string
		stateCodes := pq.Array(strings.Split(params["state-code"], ","))
		err = stmt.QueryRow(date, stateCodes).Scan(&result)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, result)
	}
}
