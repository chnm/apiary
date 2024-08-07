package apiary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// DenominationFamily describes a group of denominations. There can be different
// ways of categorizing denominations.
type DenominationFamily struct {
	Name string `json:"name"`
}

// Denomination describes a denomination's names and various systems of classification.
type Denomination struct {
	Name           string     `json:"name"`
	ShortName      string     `json:"short_name"`
	DenominationID NullString `json:"denomination_id"`
	FamilyCensus   NullString `json:"family_census"`
	FamilyRelec    string     `json:"family_relec"`
}

// RelCensusDenominationFamiliesHandler returns
func (s *Server) RelCensusDenominationFamiliesHandler() http.HandlerFunc {

	query := `
	SELECT DISTINCT family_relec 
	FROM relcensus.denominations 
	WHERE year = 1926
	ORDER BY family_relec;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]DenominationFamily, 0)
		var row DenominationFamily

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

		container := struct {
			FamilyRelec []DenominationFamily `json:"family_relec"`
		}{
			FamilyRelec: results,
		}

		response, _ := json.Marshal(container)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}
}

// RelCensusDenominationsHandler returns the denominations that are available.
// Optionally, it can be filtered to get just the denominations in a particular family.
func (s *Server) RelCensusDenominationsHandler() http.HandlerFunc {
	query := `
	SELECT denomination_id, name, short_name, family_census, family_relec
	FROM relcensus.denominations
	WHERE ($1::text = '' OR family_relec = $1::text) AND year = 1926;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		familyRelec := r.URL.Query().Get("family_relec")
		results := make([]Denomination, 0)
		var row Denomination

		rows, err := s.DB.Query(context.TODO(), query, familyRelec)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.DenominationID, &row.Name, &row.ShortName, &row.FamilyCensus, &row.FamilyRelec)
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
