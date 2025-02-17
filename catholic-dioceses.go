package apiary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// CatholicDiocese describes a diocese of the Roman Catholic Church.
type CatholicDiocese struct {
	City             string    `json:"city"`
	State            string    `json:"state"`
	Country          string    `json:"country"`
	Rite             string    `json:"rite"`
	YearErected      int64     `json:"year_erected"`
	YearMetropolitan NullInt64 `json:"year_metropolitan"`
	YearDestroyed    NullInt64 `json:"year_destroyed"`
	Lon              float32   `json:"lon"`
	Lat              float32   `json:"lat"`
}

// CatholicDiocesesPerDecade shows how many dioceses were established in North
// America per year.
type CatholicDiocesesPerDecade struct {
	Decade int64 `json:"decade"`
	Count  int64 `json:"n"`
}

// CatholicDiocesesHandler returns a JSON array of Catholic dioceses. Though
// the spatial data is stored in the database as a geometry, it is returned as
// simple lon/lat coordinates because that is easiest to process in the
// visualizations.
func (s *Server) CatholicDiocesesHandler() http.HandlerFunc {

	query := `
	SELECT city, state, country, rite, 
		date_part( 'year', date_erected) as year_erected,
		date_part('year', date_metropolitan) as year_metropolitan,
		date_part('year', date_destroyed) as year_destroyed,
		ST_X(geometry) as lon, ST_Y(geometry) as lat
	FROM catholic_dioceses
	ORDER BY date_erected;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]CatholicDiocese, 0)
		var row CatholicDiocese

		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.City, &row.State, &row.Country, &row.Rite,
				&row.YearErected, &row.YearMetropolitan, &row.YearDestroyed,
				&row.Lon, &row.Lat)
			if err != nil {
				log.Println(err)
			}
			results = append(results, row)
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
		}

		response, _ := json.Marshal(results)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}

}

// CatholicDiocesesPerDecadeHandler returns a JSON array of the number of dioceses
// established in North America per year.
func (s *Server) CatholicDiocesesPerDecadeHandler() http.HandlerFunc {

	// This query counts the number of dioceses established per decade. But it
	// also generates a series of decades from 1500 to 2020 so that there are
	// no gaps between decades in the result.
	query := `
	SELECT 
		series.decade,
		coalesce(n, 0) AS n
	FROM 
		(SELECT generate_series(1500, 2020, 10) AS decade) AS series
	LEFT JOIN 
		(SELECT 
			floor(date_part( 'year', date_erected)/10)*10 AS decade,
			count(*) AS n
		FROM catholic_dioceses
		GROUP BY decade) counts 
	ON series.decade  = counts.decade
	ORDER BY series.decade;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]CatholicDiocesesPerDecade, 0, 52) // Preallocate slice capacity
		var row CatholicDiocesesPerDecade

		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Decade, &row.Count)
			if err != nil {
				log.Println(err)
			}
			results = append(results, row)
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
		}

		response, _ := json.Marshal(results)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}

}
