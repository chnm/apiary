package apiary

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
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

// LocationInfo provides basic information about cities, counties, and states
// for use in dropdowns and other UI elements.
type LocationInfo struct {
	PlaceID    int     `json:"place_id"`
	City       string  `json:"city"`
	County     string  `json:"county"`
	State      string  `json:"state"`
	CountyAHCB string  `json:"county_ahcb"`
	MapName    string  `json:"map_name"`
	Lon        float64 `json:"lon"`
	Lat        float64 `json:"lat"`
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
		LEFT JOIN relcensus.popplaces_1926 p ON c.place_id = p.place_id
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
	WHERE m.year = $1 AND d.family_relec = $2 AND m.churches IS NOT NULL
	GROUP BY m.year, d.family_relec, m.city, m.state
	) d
	LEFT JOIN relcensus.cities_25k c ON d.city = c.city AND d.state = c.state
	LEFT JOIN relcensus.popplaces_1926 p ON c.place_id = p.place_id
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
	LEFT JOIN relcensus.popplaces_1926 p ON c.place_id = p.place_id
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
		var rows pgx.Rows

		// We've already done the error checking for the call to the API, so we can
		// just use the right query as necessary.
		switch {
		case denomination != "":
			rows, err = s.DB.Query(r.Context(), queryDenomination, yearInt, denomination)
		case denominationFamily != "":
			rows, err = s.DB.Query(r.Context(), queryFamily, yearInt, denominationFamily)
		case denomination == "" && denominationFamily == "":
			rows, err = s.DB.Query(r.Context(), queryAll, yearInt)
		}
		if err != nil {
			log.Printf("query Religious Census city membership: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var row CityMembership
			if err := rows.Scan(
				&row.Year,
				&row.Group,
				&row.City, &row.State,
				&row.Denominations, &row.Churches, &row.Members,
				&row.Population1926,
				&row.Lon, &row.Lat,
			); err != nil {
				log.Printf("scan Religious Census city membership: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			log.Printf("iterate Religious Census city membership: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(results)
		if err != nil {
			log.Printf("marshal Religious Census city membership: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(response); err != nil {
			log.Printf("write Religious Census city membership response: %v", err)
		}
	}
}

// RelCensusLocationsHandler returns a list of all locations
func (s *Server) RelCensusLocationsHandler() http.HandlerFunc {
	query := `
		SELECT DISTINCT place_id, place, county, state, county_ahcb, map_name, lat, lon
		FROM relcensus.popplaces_1926
		ORDER BY state, county, place;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]LocationInfo, 0)

		rows, err := s.DB.Query(r.Context(), query)
		if err != nil {
			log.Printf("query Religious Census locations: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var row LocationInfo
			if err := rows.Scan(
				&row.PlaceID,
				&row.City,
				&row.County,
				&row.State,
				&row.CountyAHCB,
				&row.MapName,
				&row.Lat,
				&row.Lon,
			); err != nil {
				log.Printf("scan Religious Census location: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			results = append(results, row)
		}

		if err := rows.Err(); err != nil {
			log.Printf("iterate Religious Census locations: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(results)
		if err != nil {
			log.Printf("marshal Religious Census locations: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(response); err != nil {
			log.Printf("write Religious Census locations response: %v", err)
		}
	}
}
