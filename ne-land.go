package relecapi

import (
	"fmt"
	"log"
	"net/http"
)

// NaturalEarthLandHandler returns a GeoJSON FeatureCollection containing land
// polygons from Natural Earth. The optional parameter `resolution` is passed as
// part of the query string.
func (s *Server) NaturalEarthLandHandler() http.HandlerFunc {

	// The Natural Earth land data is in three different tables, based on
	// resolution, in the `naturalearth` schema. This is because that's how the
	// shapefiles come from the Natural Earth downloads. Since the table name
	// can't be a part of a prepared statement, we need to prepare three different
	// statements for each of the tables based on which resolution we want. Then
	// we build the GeoJSON in the query itself. This query will be sent to the
	// database as a prepared statement.

	// A map of the resolutions that will be passed as query parameters and the table
	// they correspond to.
	resolutions := map[string]string{
		"10m":  "naturalearth.ne_10m_land",
		"50m":  "naturalearth.ne_50m_land",
		"110m": "naturalearth.ne_110m_land",
	}

	// Raw query with placeholder for table
	rawQuery := `
		SELECT json_build_object(
			'type','FeatureCollection',
			'features', json_agg(ne_land.feature)
		)
		FROM (
			SELECT json_build_object(
				'type', 'Feature',
				'id', gid,
				'geometry', ST_AsGeoJSON(geom)::json
			) AS feature 
		FROM %s
		) AS ne_land;
		`

	// Prepare statements for each of the tables.
	for _, v := range resolutions {
		query := fmt.Sprintf(rawQuery, v)
		stmt, err := s.Database.Prepare(query)
		if err != nil {
			log.Fatalln(err)
		}
		s.Statements[v] = stmt // Will be closed at shutdown
	}

	return func(w http.ResponseWriter, r *http.Request) {

		// Get the resolution from the query. No resolution gets the smallest data.
		// Wrong resolutions return an HTTP error.
		queryRes := r.URL.Query().Get("resolution")
		var resolution string
		switch queryRes {
		case "":
			resolution = "110m"
		case "10m", "50m", "110m":
			resolution = queryRes
		default:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var result string // result will be a string containing GeoJSON
		// Use the statement with the resolution we want.
		err := s.Statements[resolutions[resolution]].QueryRow().Scan(&result)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, result)

	}
}
