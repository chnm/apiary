package dataapi

import (
	"fmt"
	"log"
	"net/http"
)

// NENorthAmericaHandler returns a GeoJSON FeatureCollection containing country
// polygons for North America from Natural Earth.
func (s *Server) NENorthAmericaHandler() http.HandlerFunc {

	// All of the work of querying is done in this closure which is called when
	// the routes are set up. This means that the query is done only one time, at
	// startup. Essentially this a very simple cache, but it speeds up the
	// response to the client quite a bit. The downside is that if the data
	// changes in the database, the API server won't pick it up until restart.
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

	var result string // result will be a string containing GeoJSON
	err := s.Database.QueryRow(query).Scan(&result)
	if err != nil {
		log.Println(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, result)
	}

}
