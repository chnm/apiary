package relecapi

import (
	"fmt"
	"log"
	"net/http"
)

// NENorthAmericaHandler returns a GeoJSON FeatureCollection containing country
// polygons for North America from Natural Earth.
func (s *Server) NENorthAmericaHandler() http.HandlerFunc {

	query := `
		SELECT json_build_object(
			'type','FeatureCollection',
			'features', json_agg(countries.feature)
		)
		FROM (
			SELECT json_build_object(
				'type', 'Feature',
				'id', iso_a3,
				'properties', json_build_object(
					'name', name),
			  'geometry', ST_AsGeoJSON(geom_110m, 6)::json
			) AS feature
			FROM naturalearth.ne_countries
			WHERE continent = 'North America' AND iso_a3 != 'GRL'
			) AS countries;
		`
	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["ne-northamerica"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {

		var result string // result will be a string containing GeoJSON
		// Use the statement with the resolution we want.
		err := stmt.QueryRow().Scan(&result)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, result)

	}
}
