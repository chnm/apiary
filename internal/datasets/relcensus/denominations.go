package relcensus

import (
	"encoding/json"
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
func (h *Handler) RelCensusDenominationFamiliesHandler() http.HandlerFunc {

	query := `
	SELECT DISTINCT family_relec 
	FROM relcensus.denominations 
	WHERE year = 1926
	ORDER BY family_relec;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]DenominationFamily, 0)

		rows, err := h.db.Query(r.Context(), query)
		if err != nil {
			log.Printf("query Religious Census denomination families: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var row DenominationFamily
			if err := rows.Scan(&row.Name); err != nil {
				log.Printf("scan Religious Census denomination family: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			log.Printf("iterate Religious Census denomination families: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		container := struct {
			FamilyRelec []DenominationFamily `json:"family_relec"`
		}{
			FamilyRelec: results,
		}

		response, err := json.Marshal(container)
		if err != nil {
			log.Printf("marshal Religious Census denomination families: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(response); err != nil {
			log.Printf("write Religious Census denomination families response: %v", err)
		}
	}
}

// RelCensusDenominationsHandler returns the denominations that are available.
// Optionally, it can be filtered to get just the denominations in a particular family.
func (h *Handler) RelCensusDenominationsHandler() http.HandlerFunc {
	query := `
	SELECT denomination_id, name, short_name, family_census, family_relec
	FROM relcensus.denominations
	WHERE ($1::text = '' OR family_relec = $1::text) AND year = 1926;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		familyRelec := r.URL.Query().Get("family_relec")
		results := make([]Denomination, 0)

		rows, err := h.db.Query(r.Context(), query, familyRelec)
		if err != nil {
			log.Printf("query Religious Census denominations: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var row Denomination
			if err := rows.Scan(
				&row.DenominationID,
				&row.Name,
				&row.ShortName,
				&row.FamilyCensus,
				&row.FamilyRelec,
			); err != nil {
				log.Printf("scan Religious Census denomination: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			results = append(results, row)
		}
		if err := rows.Err(); err != nil {
			log.Printf("iterate Religious Census denominations: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(results)
		if err != nil {
			log.Printf("marshal Religious Census denominations: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(response); err != nil {
			log.Printf("write Religious Census denominations response: %v", err)
		}
	}

}
