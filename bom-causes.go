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

// DeathCauses returns a list of causes of death with a count of deaths for each
// cause and related metadata.
type DeathCauses struct {
	Death           string     `json:"death"`
	Count           NullInt64  `json:"count"`
	DescriptiveText NullString `json:"descriptive_text"`
	WeekID          string     `json:"week_id"`
	WeekNo          int        `json:"week_no"`
	StartDay        NullInt64  `json:"start_day"`
	StartMonth      NullString `json:"start_month"`
	EndDay          NullInt64  `json:"end_day"`
	EndMonth        NullString `json:"end_month"`
	Year            NullInt64  `json:"year"`
	SplitYear       string     `json:"split_year"`
}

// Causes describes a cause of death.
type Causes struct {
	Name string `json:"name"`
}

// DeathCausesHandler returns a JSON array of causes of death. The list of causes
// depends on whether a user has provided a comma-separated list of causes. If
// no list is provided, it returns the entire list of causes.
func (s *Server) DeathCausesHandler() http.HandlerFunc {

	queryCause := `
	SELECT 
		c.death,
		c.count, 
		c.descriptive_text, 
		c.week_id,
		w.week_no,
		w.start_day, 
		w.start_month, 
		w.end_day, 
		w.end_month, 
		y.year,
		w.split_year
	FROM 
		bom.causes_of_death c
	JOIN 
		bom.week w ON w.joinid = c.week_id
	JOIN
		bom.year y ON y.year = w.year
	WHERE 
		y.year::int >= $1
		AND y.year::int <= $2
		AND c.death = ANY($3)
		AND count IS NOT NULL
	ORDER BY 
		y.year ASC,
		w.week_no ASC,
		c.death ASC
	LIMIT $4
	OFFSET $5;
	`

	queryNoCause := `
	SELECT 
		c.death,
		c.count, 
		c.descriptive_text,
		c.week_id, 
		w.week_no,
		w.start_day, 
		w.start_month, 
		w.end_day, 
		w.end_month, 
		y.year,
		w.split_year
	FROM 
		bom.causes_of_death c
	JOIN 
		bom.week w ON w.joinid = c.week_id
	JOIN
		bom.year y ON y.year = w.year
	WHERE
		y.year::int >= $1
		AND y.year::int <= $2
		AND count IS NOT NULL
	ORDER BY 
		y.year ASC,
		w.week_no ASC,
		c.death ASC
	LIMIT $3
	OFFSET $4;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		startYear := r.URL.Query().Get("start-year")
		endYear := r.URL.Query().Get("end-year")
		causes := r.URL.Query().Get("id")
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

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

		if limit == "" {
			limit = "25"
		}
		if offset == "" {
			offset = "0"
		}

		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		causes = fmt.Sprintf("{%s}", strings.TrimSpace(causes))

		results := make([]DeathCauses, 0)
		var row DeathCauses
		var rows pgx.Rows

		switch {
		case causes == "{}":
			rows, err = s.DB.Query(context.TODO(), queryNoCause, startYearInt, endYearInt, limitInt, offsetInt)
		case causes != "{}":
			rows, err = s.DB.Query(context.TODO(), queryCause, startYearInt, endYearInt, causes, limitInt, offsetInt)
		default:
			rows, err = s.DB.Query(context.TODO(), queryNoCause, startYearInt, endYearInt, limitInt, offsetInt)
		}

		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(
				&row.Death,
				&row.Count,
				&row.DescriptiveText,
				&row.WeekID,
				&row.WeekNo,
				&row.StartDay,
				&row.StartMonth,
				&row.EndDay,
				&row.EndMonth,
				&row.Year,
				&row.SplitYear,
			)
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

func (s *Server) ListCausesHandler() http.HandlerFunc {

	query := `
	SELECT DISTINCT
		death
	FROM 
		bom.causes_of_death
	ORDER BY 
		death ASC
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]Causes, 0)
		var row Causes

		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Name)
			if err != nil {
				log.Println(err)
			}
			results = append(results, row)
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		response, _ := json.Marshal(results)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}

}
