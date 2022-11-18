package apiary

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

// SouthAmericaHandler returns a GeoJSON FeatureCollection containing country
// polygons for South America from Natural Earth.
func (s *Server) NESouthAmericaHandler() http.HandlerFunc {

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
				'id', adm0_a3,
				'properties', json_build_object(
					'name', name),
			  'geometry', ST_AsGeoJSON(geom_50m, 6)::json
			) AS feature
			FROM naturalearth.countries
			WHERE continent = 'South America' AND geom_50m IS NOT NULL
			) AS countries;
		`

	var result string // result will be a string containing GeoJSON
	err := s.DB.QueryRow(context.TODO(), query).Scan(&result)
	if err != nil {
		log.Println(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, result)
	}

}
