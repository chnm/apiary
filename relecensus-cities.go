package dataapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// DenominationInCityByYear describes the data for a specific denomination in a
// single year in a single city.
type DenominationInCityByYear struct {
	Year         int     `json:"year"`
	Denomination string  `json:"denomination"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	Churches     int     `json:"churches"`
	MembersTotal int     `json:"members_total"`
	Lon          float64 `json:"lon"`
	Lat          float64 `json:"lat"`
}

// CityMembershipHandler returns the statistics for all the cities for a single
// denomination in a single year. It must be filtered by year and denomination.
func (s *Server) CityMembershipHandler() http.HandlerFunc {
	query := `
	SELECT m.year, m.denomination, 
	  c.city, c.state,
		m.churches, m.members_total,
		ST_X(geometry) AS lon, ST_Y(geometry) AS lat
	FROM relcensus.cities_25K c
	LEFT JOIN relcensus.membership_city m
	ON c.city = m.city AND c.state = m.state
	WHERE year = $1 AND denomination = $2
	ORDER BY state, city;
	`
	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["denomination-in-city"] = stmt
	return func(w http.ResponseWriter, r *http.Request) {
		year := r.URL.Query().Get("year")
		denomination := r.URL.Query().Get("denomination")

		if year == "" || denomination == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		results := make([]DenominationInCityByYear, 0)
		var row DenominationInCityByYear

		rows, err := stmt.Query(year, denomination)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Year, &row.Denomination, &row.City, &row.State,
				&row.Churches, &row.MembersTotal, &row.Lon, &row.Lat)
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
		fmt.Fprintf(w, string(response))
	}

}
