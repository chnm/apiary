package apiary

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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

// BillsShapefilesHandler returns a GeoJSON FeatureCollection containing parish
// polygons joined with the bills data. It accepts filtering by year, bill_type,
// count_type, etc.
func (s *Server) BillsShapefilesHandler() http.HandlerFunc {
	// Base query with materialized CTE and spatial index hints for performance
	baseQuery := `
    WITH filtered_bills AS MATERIALIZED (
        SELECT 
            b.parish_id,
            b.count_type,
            b.count,
            b.year
        FROM 
            bom.bill_of_mortality b
        WHERE 1=1
        -- Dynamic bill filters will be added here
    ),
    parish_data AS (
        SELECT 
            parishes_shp.id,
            parishes_shp.par,
            parishes_shp.civ_par,
            parishes_shp.dbn_par,
            parishes_shp.omeka_par,
            parishes_shp.subunit,
            parishes_shp.city_cnty,
            parishes_shp.start_yr,
            parishes_shp.sp_total,
            parishes_shp.sp_per,
            COALESCE(SUM(CASE WHEN fb.count_type = 'Buried' THEN fb.count ELSE 0 END), 0) as total_buried,
            COALESCE(SUM(CASE WHEN fb.count_type = 'Plague' THEN fb.count ELSE 0 END), 0) as total_plague,
            COUNT(fb.parish_id) as bill_count,
            parishes_shp.geom_01
        FROM 
            bom.parishes_shp
				LEFT JOIN
      		filtered_bills fb ON fb.parish_id = parishes_shp.parish_id
        WHERE 1=1
        -- Dynamic parish filters will be added here
        GROUP BY
            parishes_shp.id, parishes_shp.par, parishes_shp.civ_par, parishes_shp.dbn_par,
            parishes_shp.omeka_par, parishes_shp.subunit, parishes_shp.city_cnty,
            parishes_shp.start_yr, parishes_shp.sp_total, parishes_shp.sp_per, parishes_shp.geom_01
    )
    SELECT json_build_object(
        'type', 'FeatureCollection',
        'features', COALESCE(json_agg(features.feature), '[]'::json)
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
                'sp_per', sp_per,
                'total_buried', total_buried,
                'total_plague', total_plague,
                'bill_count', bill_count
            ),
            'geometry', ST_AsGeoJSON(
                ST_Transform(
                    ST_SetSRID(geom_01, 27700), 
                    4326
                ), 
                6
            )::json
        ) AS feature
        FROM parish_data
    ) AS features;
    `

	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		year := r.URL.Query().Get("year")
		startYear := r.URL.Query().Get("start-year")
		endYear := r.URL.Query().Get("end-year")
		subunit := r.URL.Query().Get("subunit")
		cityCounty := r.URL.Query().Get("city_cnty")
		billType := r.URL.Query().Get("bill-type")
		countType := r.URL.Query().Get("count-type")
		parish := r.URL.Query().Get("parish")

		// Build the query with separate filters for bills and parishes
		billFilters, parishFilters := buildSeparateFilters(
			year, startYear, endYear, subunit, cityCounty, billType, countType, parish)

		// Apply the filters to their respective sections
		query := strings.Replace(baseQuery, "-- Dynamic bill filters will be added here", billFilters, 1)
		query = strings.Replace(query, "-- Dynamic parish filters will be added here", parishFilters, 1)

		// Execute query with a timeout context to prevent long-running queries
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var result string
		err := s.DB.QueryRow(ctx, query).Scan(&result)
		if err != nil {
			log.Printf("Error executing bills shapefile query: %v", err)
			// Check for context deadline exceeded to provide better error messaging
			if ctx.Err() == context.DeadlineExceeded {
				log.Printf("Query timed out, consider optimizing or using more specific filters")
				http.Error(w, "Query timed out. Please try with more specific filters.", http.StatusRequestTimeout)
				return
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Set appropriate headers for GeoJSON response with optimized caching
		w.Header().Set("Content-Type", "application/geo+json")
		w.Header().Set("Cache-Control", "public, max-age=86400") // 24 hours cache
		w.Header().Set("Vary", "Accept-Encoding")                // Allow caching of different encodings
		fmt.Fprint(w, result)
	}
}

// buildSeparateFilters constructs separate SQL filters for bills and parishes based on URL parameters
func buildSeparateFilters(year, startYear, endYear, subunit, cityCounty, billType, countType, parish string) (string, string) {
	var billFilters []string
	var parishFilters []string

	// Add filters based on provided parameters
	if year != "" {
		if yearInt, err := strconv.Atoi(year); err == nil {
			billFilters = append(billFilters, fmt.Sprintf("AND b.year = %d", yearInt))
			parishFilters = append(parishFilters, fmt.Sprintf("AND parishes_shp.start_yr = %d", yearInt))
		}
	} else {
		// Use start-year and end-year if provided
		if startYear != "" {
			if startYearInt, err := strconv.Atoi(startYear); err == nil {
				billFilters = append(billFilters, fmt.Sprintf("AND b.year >= %d", startYearInt))
			}
		}
		if endYear != "" {
			if endYearInt, err := strconv.Atoi(endYear); err == nil {
				billFilters = append(billFilters, fmt.Sprintf("AND b.year <= %d", endYearInt))
			}
		}
	}

	// Parish-specific filters
	if subunit != "" {
		parishFilters = append(parishFilters, fmt.Sprintf("AND parishes_shp.subunit = '%s'", subunit))
	}

	if cityCounty != "" {
		parishFilters = append(parishFilters, fmt.Sprintf("AND parishes_shp.city_cnty = '%s'", cityCounty))
	}

	// Bills-specific filters
	if billType != "" && IsValidBillType(billType) {
		billFilters = append(billFilters, fmt.Sprintf("AND b.bill_type = '%s'", billType))
	}

	if countType != "" && IsValidCountType(countType) {
		billFilters = append(billFilters, fmt.Sprintf("AND b.count_type = '%s'", countType))
	}

	// Add parish filter to both queries to ensure they're properly joined
	if parish != "" {
		parishIDs := strings.Split(parish, ",")
		var validParishIDs []string

		for _, id := range parishIDs {
			if trimmedID := strings.TrimSpace(id); trimmedID != "" {
				if _, err := strconv.Atoi(trimmedID); err == nil {
					validParishIDs = append(validParishIDs, trimmedID)
				}
			}
		}

		if len(validParishIDs) > 0 {
			parishFilter := fmt.Sprintf("AND parishes_shp.id IN (%s)", strings.Join(validParishIDs, ","))
			parishFilters = append(parishFilters, parishFilter)
			billFilter := fmt.Sprintf("AND b.parish_id IN (%s)", strings.Join(validParishIDs, ","))
			billFilters = append(billFilters, billFilter)
		}
	}

	return strings.Join(billFilters, " "), strings.Join(parishFilters, " ")
}
