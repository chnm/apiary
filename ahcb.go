package relecapi

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"log"
	"net/http"
	"strings"
	"time"
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

// AHCBCountiesHandler returns a GeoJSON FeatureCollection containing countries
// from AHCB. The handler will get the county boundaries for a particular date.
// Optionally, a comma-separated array of values for either county IDs or the
// state/territory names can be used to filter the results.
func (s *Server) AHCBCountiesHandler() http.HandlerFunc {

	// The minimum and maximum dates are the range of dates for states in AHCB.
	minDate, _ := time.Parse("2006-01-02", "1629-03-04")
	maxDate, _ := time.Parse("2006-01-02", "2000-12-31")

	// Build the GeoJSON in the query itself. This query will be sent to the
	// database as a prepared statement.
	query := `
		SELECT json_build_object(
			'type','FeatureCollection',
			'features', json_agg(us_counties.feature)
		)
		FROM (
			SELECT json_build_object(
				'type', 'Feature',
				'id', id,
				'geometry', ST_AsGeoJSON(geom_01)::json,
				'properties', json_build_object(
					'name', name,
					'state_terr', state_terr,
					'area_sqmi', area_sqmi)
			) AS feature
			FROM ahcb_counties
			WHERE start_date <= $1 AND end_date >= $1
			) AS us_counties;
		`
	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["ahcb-counties"] = stmt // Will be closed at shutdown

	queryStateterrFilter := `
		SELECT json_build_object(
			'type','FeatureCollection',
			'features', json_agg(us_counties.feature)
		)
		FROM (
			SELECT json_build_object(
				'type', 'Feature',
				'id', id,
				'geometry', ST_AsGeoJSON(geom_01)::json,
				'properties', json_build_object(
					'name', name,
					'state_terr', state_terr,
					'area_sqmi', area_sqmi)
			) AS feature
			FROM ahcb_counties
			WHERE start_date <= $1 AND end_date >= $1
			AND state_terr = ANY($2)
			) AS us_counties;
		`
	stmtStateterrFilter, err := s.Database.Prepare(queryStateterrFilter)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["ahcb-counties-stateterr-filter"] = stmtStateterrFilter // Will be closed at shutdown

	queryCountyFilter := `
		SELECT json_build_object(
			'type','FeatureCollection',
			'features', json_agg(us_counties.feature)
		)
		FROM (
			SELECT json_build_object(
				'type', 'Feature',
				'id', id,
				'geometry', ST_AsGeoJSON(geom_01)::json,
				'properties', json_build_object(
					'name', name,
					'state_terr', state_terr,
					'area_sqmi', area_sqmi)
			) AS feature
			FROM ahcb_counties
			WHERE start_date <= $1 AND end_date >= $1
			AND id = ANY($2)
			) AS us_counties;
		`
	stmtCountyFilter, err := s.Database.Prepare(queryCountyFilter)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["ahcb-counties-county-filter"] = stmtCountyFilter // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		date, err := dateInRange(params["date"], minDate, maxDate)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var result string // result will be a string containing GeoJSON
		stateterrValue := r.FormValue("state_terr")
		idValue := r.FormValue("id")
		if "" != stateterrValue {
			stateterrs := pq.Array(strings.Split(stateterrValue, ","))
			err = stmtStateterrFilter.QueryRow(date, stateterrs).Scan(&result)
		} else if "" != idValue {
			ids := pq.Array(strings.Split(idValue, ","))
			err = stmtCountyFilter.QueryRow(date, ids).Scan(&result)
		} else {
			err = stmt.QueryRow(date).Scan(&result)
		}
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, result)

	}
}
