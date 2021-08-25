package dataapi

import (
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
	Name           string `json:"name"`
	DenominationID string `json:"denomination_id"`
	FamilyCensus   string `json:"family_census"`
	FamilyARDA     string `json:"family_arda"`
	FamilyRelec    string `json:"family_relec"`
	ID             string `json:"id"`
}

// DenominationFamiliesHandler returns
func (s *Server) DenominationFamiliesHandler() http.HandlerFunc {

	query := `
	SELECT DISTINCT family_relec 
	FROM relcensus.denominations 
	ORDER BY family_relec;
	`

	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["denomination-families"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]DenominationFamily, 0)
		var row DenominationFamily

		rows, err := stmt.Query()
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
		fmt.Fprintf(w, string(response))
	}
}
