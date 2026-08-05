package catholic

import (
	"encoding/json"
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
func (h *Handler) CatholicDiocesesHandler() http.HandlerFunc {

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

		rows, err := h.db.Query(r.Context(), query)
		if err != nil {
			log.Printf("query Catholic dioceses: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var row CatholicDiocese
			if err := rows.Scan(&row.City, &row.State, &row.Country, &row.Rite,
				&row.YearErected, &row.YearMetropolitan, &row.YearDestroyed,
				&row.Lon, &row.Lat); err != nil {
				log.Printf("scan Catholic diocese: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			log.Printf("iterate Catholic dioceses: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(results)
		if err != nil {
			log.Printf("marshal Catholic dioceses: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(response); err != nil {
			log.Printf("write Catholic dioceses response: %v", err)
		}
	}

}

// CatholicDiocesesPerDecadeHandler returns a JSON array of the number of dioceses
// established in North America per year.
func (h *Handler) CatholicDiocesesPerDecadeHandler() http.HandlerFunc {

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

		rows, err := h.db.Query(r.Context(), query)
		if err != nil {
			log.Printf("query Catholic dioceses per decade: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var row CatholicDiocesesPerDecade
			if err := rows.Scan(&row.Decade, &row.Count); err != nil {
				log.Printf("scan Catholic dioceses per decade: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			log.Printf("iterate Catholic dioceses per decade: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(results)
		if err != nil {
			log.Printf("marshal Catholic dioceses per decade: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(response); err != nil {
			log.Printf("write Catholic dioceses per decade response: %v", err)
		}
	}

}
