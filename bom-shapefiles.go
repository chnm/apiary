package apiary

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// ParishShpHandler returns a GeoJSON FeatureCollection containing parish
// polygons and can receive year, city_cnty, and subunit parameters.
func (s *Server) ParishShpHandler() http.HandlerFunc {
	baseQuery := `
      SELECT json_build_object(
        'type', 'FeatureCollection',
        'features', COALESCE(json_agg(parishes.feature), '[]'::json)
    )
    FROM (
        SELECT json_build_object(
            'type', 'Feature',
            'id', id,
            'properties', json_build_object(
                'par', par,
                'civ_par', civ_par,
                'dbn_par', dbn_par,
                'omeka_par', omeka_par,
                'subunit', subunit,
                'city_cnty', city_cnty,
                'start_yr', start_yr,
                'sp_total', sp_total,
                'sp_per', sp_per
            ),
            'geometry', ST_AsGeoJSON(
                ST_Transform(
                    ST_SetSRID(geom_01, 27700), 
                    4326
                ), 
                6
            )::json
        ) AS feature
        FROM bom.parishes_shp
        WHERE 1=1
      `

	return func(w http.ResponseWriter, r *http.Request) {
		// Get query parameters
		year := r.URL.Query().Get("year")
		subunit := r.URL.Query().Get("subunit")
		cityCounty := r.URL.Query().Get("city_cnty")

		// Build the query with optional filters
		query := baseQuery
		var params []interface{}
		paramCount := 1

		if year != "" {
			if yearInt, err := strconv.Atoi(year); err == nil {
				query += fmt.Sprintf(" AND start_yr = $%d", paramCount)
				params = append(params, yearInt)
				paramCount++
			}
		}

		if subunit != "" {
			query += fmt.Sprintf(" AND subunit = $%d", paramCount)
			params = append(params, subunit)
			paramCount++
		}

		if cityCounty != "" {
			query += fmt.Sprintf(" AND city_cnty = $%d", paramCount)
			params = append(params, cityCounty)
			paramCount++
		}

		// Close the subquery and main query
		query += ") AS parishes;"

		var result string
		var err error

		if len(params) == 0 {
			err = s.DB.QueryRow(context.Background(), query).Scan(&result)
		} else {
			err = s.DB.QueryRow(context.Background(), query, params...).Scan(&result)
		}

		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, result)
	}
}
