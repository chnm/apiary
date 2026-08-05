package presbyterians

import (
	"encoding/json"
	"log"
	"net/http"
)

// PresbyteriansByYear holds aggregate data on Presbyterian membership and churches.
type PresbyteriansByYear struct {
	Year     int `json:"year"`
	Members  int `json:"members"`
	Churches int `json:"churches"`
}

// PresbyteriansHandler returns the aggregate data on Presbyterian memberhsip and churches.
func (h *Handler) PresbyteriansHandler() http.HandlerFunc {

	query := `
	SELECT 
		year, 
		SUM(members) as members, 
		SUM(churches) as churches
	FROM presbyterians_weber 
	WHERE members IS NOT NULL 
	GROUP BY year 
	ORDER BY year;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]PresbyteriansByYear, 0)

		rows, err := h.db.Query(r.Context(), query)
		if err != nil {
			log.Printf("query Presbyterian statistics: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var row PresbyteriansByYear
			if err := rows.Scan(&row.Year, &row.Members, &row.Churches); err != nil {
				log.Printf("scan Presbyterian statistics: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			log.Printf("iterate Presbyterian statistics: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(results)
		if err != nil {
			log.Printf("marshal Presbyterian statistics: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(response); err != nil {
			log.Printf("write Presbyterian statistics response: %v", err)
		}
	}
}
