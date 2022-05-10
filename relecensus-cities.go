package apiary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v4"
)

// CityMembership gives the membership (and population) statistics for some
// aggregation of denominations in a given year.
type CityMembership struct {
	Year           int       `json:"year"`
	Group          string    `json:"group"`
	City           string    `json:"city"`
	State          string    `json:"state"`
	Denominations  int       `json:"denominations"`
	Churches       int       `json:"churches"`
	Members        int       `json:"members"`
	Population1926 NullInt64 `json:"population_1926"`
	Lon            float64   `json:"lon"`
	Lat            float64   `json:"lat"`
}

// RelCensusCityMembershipHandler returns the statistics for all the cities for a single
// denomination in a single year. It must be filtered by year and denomination.
func (s *Server) RelCensusCityMembershipHandler() http.HandlerFunc {
	queryDenomination := `
		SELECT m.year, m.denomination, 
		c.city, c.state,
		1::integer AS denominations,
		m.churches, m.members_total,
		p.pop_est_1926,
		ST_X(c.geometry) AS lon, ST_Y(c.geometry) AS lat
		FROM relcensus.membership_city m
		LEFT JOIN relcensus.cities_25K c ON m.city = c.city AND m.state = c.state
		LEFT JOIN popplaces_1926 p ON c.place_id = p.place_id
		WHERE year = $1 AND denomination = $2
		ORDER BY state, city;
	`

	queryFamily := `
	SELECT 
	d.year,
	d.family_relec,
	c.city, c.state, 
	d.denominations, 
	d.churches, 
	d.members_total, 
	p.pop_est_1926,
	ST_X(c.geometry) AS lon, ST_Y(c.geometry) AS lat
	FROM
	(
	SELECT 
	m.year, 
	d.family_relec, 
	m.city, m.state,
	count(m.denomination) AS denominations, 
	sum(m.churches) AS churches, 
	sum(m.members_total) AS members_total
	FROM relcensus.membership_city m
	LEFT JOIN relcensus.denominations d ON m.denomination = d.name
	WHERE m.year = $1 AND d.family_relec = $2
	GROUP BY m.year, d.family_relec, m.city, m.state
	) d
	LEFT JOIN relcensus.cities_25k c ON d.city = c.city AND d.state = c.state
	LEFT JOIN popplaces_1926 p ON c.place_id = p.place_id
	ORDER BY c.state, c.city;
	`

	queryAll := `
	SELECT 
	d.year,
	'All denominations' AS group,
	c.city, c.state, 
	d.denominations, 
	d.churches, 
	d.members_total, 
	p.pop_est_1926,
	ST_X(c.geometry) AS lon, ST_Y(c.geometry) AS lat
	FROM
	(
	SELECT 
		m.year, 
		m.city, m.state,
		count(m.denomination) AS denominations, 
		sum(m.churches) AS churches, 
		sum(m.members_total) AS members_total
	FROM relcensus.membership_city m
	WHERE m.year = $1 
	GROUP BY m.year, m.city, m.state
	) d
	LEFT JOIN relcensus.cities_25k c ON d.city = c.city AND d.state = c.state
	LEFT JOIN popplaces_1926 p ON c.place_id = p.place_id
	ORDER BY c.state, c.city;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		year := r.URL.Query().Get("year")
		denomination := r.URL.Query().Get("denomination")
		denominationFamily := r.URL.Query().Get("denominationFamily")

		// Year must be provided
		if year == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// Year must be an integer
		yearInt, err := strconv.Atoi(year)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// Year must be one of the following
		switch yearInt {
		case 1906:
		case 1916:
		case 1926:
		case 1936:
		default:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// Only allow one of denomination or denominationFamily to be set
		if denomination != "" && denominationFamily != "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		results := make([]CityMembership, 0)
		var row CityMembership
		var rows pgx.Rows

		// We've already done the error checking for the call to the API, so we can
		// just use the right query as necessary.
		switch {
		case denomination != "":
			rows, err = s.DB.Query(context.TODO(), queryDenomination, yearInt, denomination)
		case denominationFamily != "":
			rows, err = s.DB.Query(context.TODO(), queryFamily, yearInt, denominationFamily)
		case denomination == "" && denominationFamily == "":
			rows, err = s.DB.Query(context.TODO(), queryAll, yearInt)
		}
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(
				&row.Year,
				&row.Group,
				&row.City, &row.State,
				&row.Denominations, &row.Churches, &row.Members,
				&row.Population1926,
				&row.Lon, &row.Lat)
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
