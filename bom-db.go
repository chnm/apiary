package dataapi

import (
	// 	"encoding/json"
	// 	"fmt"
	// 	"log"
	"log"
	"net/http"
	// 	"strconv"
)

// ParishByWeek describes the data for a specific parish in
// a given week.
type ParishByWeek struct {
	Year    int     `json:"year"`
	Parish  string  `json:"parish"`
	Deaths  int     `json:"deaths"`
	Burials int     `json:"burials"`
	Lon     float64 `json:"lon"`
	Lat     float64 `json:"lat"`
}

// WeekHandler returns the statistics for all the weeks for
// a single parish in a single week. It is filtered by year and parish name.
func (s *Server) WeekHandler() http.HandlerFunc {
	query := `
		SELECT year, parish, deaths, burials, 
		ST_X(geometry) AS lon, ST_Y(geometry) AS lat
		FROM bom.bom_db 
		WHERE week = $1 AND year = $2 AND parish = $3
	`

	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	s.Statements["totals-week"] = stmt

	// . . .
}
