package relecapi

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
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
