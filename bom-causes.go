package apiary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
)

// DeathCauses TODO: Describe
type DeathCauses struct {
	DeathID    int        `json:"death_id"`
	Death      string     `json:"death"`
	Count      NullInt64  `json:"count"`
	WeekID     string     `json:"week_id"`
	StartDay   NullInt64  `json:"start_day"`
	StartMonth NullString `json:"start_month"`
	EndDay     NullInt64  `json:"end_day"`
	EndMonth   NullString `json:"end_month"`
	Year       int        `json:"year"`
	SplitYear  string     `json:"split_year"`
}

// DeathCausesHandler TODO: Describe
func (s *Server) DeathCausesHandler() http.HandlerFunc {

	queryCause := `
	SELECT 
		c.id, 
		c.death, 
		c.count, 
		c.week_id, 
		w.start_day, 
		w.start_month, 
		w.end_day, 
		w.end_month, 
		y.year,
		y.split_year
	FROM 
		bom.causes_of_death c
	JOIN 
		bom.week w ON w.week_id = c.week_id
	JOIN
		bom.year y ON y.year_id = w.year_id
	WHERE 
		y.year >= $1
		AND y.year <= $2
		AND (
			$3::text[] IS NULL
			OR c.death = ANY($3::text[])
		)
	ORDER BY 
		y.year ASC,
		c.death ASC;
	`

	queryNoCause := `
	SELECT 
		c.id, 
		c.death, 
		c.count, 
		c.week_id, 
		w.start_day, 
		w.start_month, 
		w.end_day, 
		w.end_month, 
		y.year,
		y.split_year
	FROM 
		bom.causes_of_death c
	JOIN 
		bom.week w ON w.week_id = c.week_id
	JOIN
		bom.year y ON y.year_id = w.year_id
	WHERE 
		y.year >= $1
		AND y.year <= $2
	ORDER BY 
		y.year ASC,
		c.death ASC;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		startYear := r.URL.Query().Get("start-year")
		endYear := r.URL.Query().Get("end-year")
		causes := r.URL.Query().Get("causes")

		if startYear == "" || endYear == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		startYearInt, err := strconv.Atoi(startYear)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		endYearInt, err := strconv.Atoi(endYear)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// causes needs to be a postgres array of integers to send to the query. This
		// returns '{1, 2, 3}' to give the query a literal array.
		causes = fmt.Sprintf("{%s}", strings.TrimSpace(causes))

		results := make([]DeathCauses, 0)
		var row DeathCauses
		var rows pgx.Rows

		switch {
		case causes == "{}":
			rows, err = s.DB.Query(context.TODO(), queryNoCause, startYearInt, endYearInt)
		case causes != "{}":
			rows, err = s.DB.Query(context.TODO(), queryCause, startYearInt, endYearInt, causes)
		default:
			rows, err = s.DB.Query(context.TODO(), queryNoCause, startYearInt, endYearInt)
		}

		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(
				&row.DeathID,
				&row.Death,
				&row.Count,
				&row.WeekID,
				&row.StartDay,
				&row.StartMonth,
				&row.EndDay,
				&row.EndMonth,
				&row.Year,
				&row.SplitYear)
			if err != nil {
				log.Println(err)
			}
			results = append(results, row)

		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, _ := json.Marshal(results)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}
}
