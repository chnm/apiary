package dataapi

// "encoding/json"
// "fmt"
"log"
// "net/http"

// BoMParishes describes the parishes.
type BoMParishes struct {
	Name string `json:"name"`
}

// Parish describes a parish's name and parish collective.
type Parish struct {
	Name       string `json:"name"`
	Collective string `json:"collective"`
	ID         string `json:"id"`
}

// ParishesHandler returns...
func (s *Server) ParishesHandler() http.HandlerFunc {
	
	query := `
	SELECT DISTINCT name
	FROM parishes
	ORDER BY name
	`

	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	s.Statements["parishes"] = stmt

	// . . .
}