package bom

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// BillsShapefilesHandler returns a GeoJSON FeatureCollection containing parish
// polygons joined with the bills data. It accepts filtering by year, bill_type,
// count_type, etc. Malformed filter values return 400 Bad Request.
func (h *Handler) BillsShapefilesHandler() http.HandlerFunc {
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
    unique_parishes AS (
        SELECT DISTINCT ON (civ_par, start_yr, ST_AsText(geom_01))
            id,
            par,
            civ_par,
            dbn_par,
            omeka_par,
            subunit,
            city_cnty,
            start_yr,
            sp_total,
            sp_per,
            geom_01
        FROM bom.parishes_shp
        WHERE 1=1
        -- Dynamic parish filters will be added here
    ),
    parish_data AS (
        SELECT
            unique_parishes.id,
            unique_parishes.par,
            unique_parishes.civ_par,
            unique_parishes.dbn_par,
            unique_parishes.omeka_par,
            unique_parishes.subunit,
            unique_parishes.city_cnty,
            unique_parishes.start_yr,
            unique_parishes.sp_total,
            unique_parishes.sp_per,
            COALESCE(SUM(CASE WHEN fb.count_type = 'buried' THEN fb.count ELSE 0 END), 0) as total_buried,
            COALESCE(SUM(CASE WHEN fb.count_type = 'plague' THEN fb.count ELSE 0 END), 0) as total_plague,
            COUNT(fb.parish_id) as bill_count,
            unique_parishes.geom_01
        FROM
            unique_parishes
				LEFT JOIN
      		bom.parishes p ON LOWER(REPLACE(REPLACE(p.canonical_name, '-', ' '), '.', '')) = LOWER(REPLACE(REPLACE(unique_parishes.civ_par, '-', ' '), '.', ''))
				LEFT JOIN
      		filtered_bills fb ON fb.parish_id = p.id
        WHERE 1=1
        -- Dynamic parish filters will be added here
        GROUP BY
            unique_parishes.id, unique_parishes.par, unique_parishes.civ_par, unique_parishes.dbn_par,
            unique_parishes.omeka_par, unique_parishes.subunit, unique_parishes.city_cnty,
            unique_parishes.start_yr, unique_parishes.sp_total, unique_parishes.sp_per, unique_parishes.geom_01
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
		billFilters, parishFilters, params, err := buildSeparateFilters(
			year, startYear, endYear, subunit, cityCounty, billType, countType, parish)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Apply the filters to their respective sections
		query := strings.Replace(baseQuery, "-- Dynamic bill filters will be added here", billFilters, 1)
		query = strings.Replace(query, "-- Dynamic parish filters will be added here", parishFilters, 1)

		// Execute query with a timeout context to prevent long-running queries
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		var result string
		err = h.db.QueryRow(ctx, query, params...).Scan(&result)
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

// buildSeparateFilters constructs parameterized SQL filters for bills and
// parishes based on URL parameters.
func buildSeparateFilters(year, startYear, endYear, subunit, cityCounty, billType, countType, parish string) (string, string, []any, error) {
	var billFilters []string
	var parishFilters []string
	var params []any

	addParam := func(value any) string {
		params = append(params, value)
		return fmt.Sprintf("$%d", len(params))
	}

	// Add filters based on provided parameters
	// Note: Year filters only apply to bills, not parish geometries
	// Parish geometries are filtered by other attributes (subunit, city_cnty, parish ID)
	if year != "" {
		yearInt, err := strconv.Atoi(year)
		if err != nil {
			return "", "", nil, errors.New("year must be an integer")
		}
		billFilters = append(billFilters, fmt.Sprintf("AND b.year = %s", addParam(yearInt)))
	} else {
		// Use start-year and end-year if provided
		if startYear != "" {
			startYearInt, err := strconv.Atoi(startYear)
			if err != nil {
				return "", "", nil, errors.New("start-year must be an integer")
			}
			billFilters = append(billFilters, fmt.Sprintf("AND b.year >= %s", addParam(startYearInt)))
		}
		if endYear != "" {
			endYearInt, err := strconv.Atoi(endYear)
			if err != nil {
				return "", "", nil, errors.New("end-year must be an integer")
			}
			billFilters = append(billFilters, fmt.Sprintf("AND b.year <= %s", addParam(endYearInt)))
		}
	}

	// Parish-specific filters
	if subunit != "" {
		parishFilters = append(parishFilters, fmt.Sprintf("AND parishes_shp.subunit = %s", addParam(subunit)))
	}

	if cityCounty != "" {
		parishFilters = append(parishFilters, fmt.Sprintf("AND parishes_shp.city_cnty = %s", addParam(cityCounty)))
	}

	// Bills-specific filters
	if billType != "" {
		if !IsValidBillType(billType) {
			return "", "", nil, errors.New("invalid bill-type")
		}
		billFilters = append(billFilters, fmt.Sprintf("AND b.bill_type = %s", addParam(strings.ToLower(billType))))
	}

	if countType != "" {
		if !IsValidCountType(countType) {
			return "", "", nil, errors.New("invalid count-type")
		}
		billFilters = append(billFilters, fmt.Sprintf("AND b.count_type = %s", addParam(strings.ToLower(countType))))
	}

	// Add parish filter to both queries to ensure they're properly joined
	if parish != "" {
		parishIDs := strings.Split(parish, ",")
		validParishIDs := make([]int, 0, len(parishIDs))

		for _, id := range parishIDs {
			trimmedID := strings.TrimSpace(id)
			parishID, err := strconv.Atoi(trimmedID)
			if err != nil || parishID <= 0 {
				return "", "", nil, errors.New("invalid parish ID")
			}
			validParishIDs = append(validParishIDs, parishID)
		}

		parishIDsParam := addParam(validParishIDs)
		parishFilters = append(parishFilters, fmt.Sprintf("AND parishes_shp.id = ANY(%s)", parishIDsParam))
		billFilters = append(billFilters, fmt.Sprintf("AND b.parish_id = ANY(%s)", parishIDsParam))
	}

	return strings.Join(billFilters, " "), strings.Join(parishFilters, " "), params, nil
}
