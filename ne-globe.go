package apiary

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

// NaturalEarthHandler returns a GeoJSON FeatureCollection containing country
// polygons by passing location parameters.
// The available country parameters are:
// Africa; Antarctica; Asia; Europe; North+America; Oceania; South+America; Seven+seas+(open+ocean)
func (s *Server) NaturalEarthHandler() http.HandlerFunc {

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
			) AS countries;
		`

	return func(w http.ResponseWriter, r *http.Request) {
		location := r.URL.Query()["location"]
		var result string

		// If no location is provided, return all countries.
		if len(location) == 0 {
			err := s.DB.QueryRow(context.Background(), query).Scan(&result)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, result)
			return
		}

		// If multiple location values are provided (e.g., ?location=Europe&location=Asia)
		// then the query will return a FeatureCollection with all of the countries
		// from each continent.
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
				WHERE continent = ANY($1) AND geom_50m IS NOT NULL
			) AS countries;
		`

		err := s.DB.QueryRow(context.Background(), query, location).Scan(&result)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, result)
	}
}
