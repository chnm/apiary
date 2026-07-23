package apiary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Parish describes a parish name, canonical name, and unique ID.
type Parish struct {
	ParishID       int        `json:"id"`
	Name           string     `json:"name"`
	CanonicalName  string     `json:"canonical_name"`
	BillSubunit    NullString `json:"subunit"`
	FoundationYear NullString `json:"foundation_year"`
	Notes          NullString `json:"notes"`
}

// missingIDs returns the members of requested that are not present in found,
// in request order and without duplicates. A nil result means nothing is missing.
func missingIDs(requested []int, found map[int]bool) []int {
	var missing []int
	seen := make(map[int]bool, len(requested))
	for _, id := range requested {
		if !found[id] && !seen[id] {
			missing = append(missing, id)
			seen[id] = true
		}
	}
	return missing
}

// invalidParishIDs returns the parish IDs from ids that do not exist in the
// bom.parishes table. A nil result means all IDs are valid. An error is
// returned only if the lookup query itself fails.
func (s *Server) invalidParishIDs(ctx context.Context, ids []int) ([]int, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `SELECT id FROM bom.parishes WHERE id = ANY($1);`
	rows, err := s.DB.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	found := make(map[int]bool, len(ids))
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		found[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return missingIDs(ids, found), nil
}

// ParishesHandler returns a list of unique parish IDs and names.
func (s *Server) ParishesHandler() http.HandlerFunc {
	query := `
	SELECT id, parish_name, canonical_name, bills_subunit, foundation_year, notes
	FROM bom.parishes
	ORDER BY canonical_name;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]Parish, 0)
		var row Parish

		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.ParishID, &row.Name, &row.CanonicalName, &row.BillSubunit, &row.FoundationYear, &row.Notes)
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
