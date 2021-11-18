package dataapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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

// CityMembershipHandler returns the statistics for all the cities for a single
// denomination in a single year. It must be filtered by year and denomination.
func (s *Server) CityMembershipHandler() http.HandlerFunc {
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
	stmtDenomination, err := s.Database.Prepare(queryDenomination)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["city-denomination"] = stmtDenomination

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
	LEFT JOIN relcensus.denominations d ON m.denomination_id = d.denomination_id
	WHERE m.year = $1 AND d.family_relec = $2
	GROUP BY m.year, d.family_relec, m.city, m.state
	) d
	LEFT JOIN relcensus.cities_25k c ON d.city = c.city AND d.state = c.state
	LEFT JOIN popplaces_1926 p ON c.place_id = p.place_id;
	`
	stmtFamily, err := s.Database.Prepare(queryFamily)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["city-family"] = stmtFamily

	queryAll := `
	// SELECT 
	// 	year, 
	// 	'All denominations' AS group,
	// 	city,
	// 	state,
	// 	denominations,
	// 	churches,
	// 	members_total,
	// 	population_1926,
	//   ST_X(geometry) AS lon,
	// 	ST_Y(geometry) AS lat
	// FROM relcensus.membership_totals_city
	// WHERE year = $1;
	`
	stmtAll, err := s.Database.Prepare(queryAll)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["city-all-denominations"] = stmtAll

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
		switch {
		case denomination == "" && denominationFamily == "":
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		case denomination != "" && denominationFamily != "":
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}

		results := make([]CityMembership, 0)
		var row CityMembership
		var rows *sql.Rows

		// We've already done the error checking for the call to the API, so we can
		// just use the right query as necessary.
		switch {
		case denomination != "":
			rows, err = stmtDenomination.Query(yearInt, denomination)
		case denominationFamily != "":
			rows, err = stmtFamily.Query(yearInt, denominationFamily)
		case denomination == "" && denominationFamily == "":
			rows, err = stmtAll.Query(yearInt)
		}
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(
				&row.Year, &row.Group, &row.City, &row.State,
				&row.Denominations, &row.Churches, &row.Members,
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
		fmt.Fprintf(w, string(response))
	}

}
