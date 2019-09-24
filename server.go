package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // Driver for database
	"os"
)

// The Server type shares access to the database.
type Server struct {
	Database *sql.DB
	Router   *mux.Router
}

// NewServer creates a new Server and connects to the database or fails trying.
func NewServer() (*Server, error) {
	s := Server{}

	// Connect to the database
	constr := fmt.Sprintf("user=releco dbname=releco password=%s host=localhost port=5555 sslmode=disable", os.Getenv("DBPASS"))
	db, err := sql.Open("postgres", constr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create the router
	router := mux.NewRouter()

	s.Database = db
	s.Router = router

	return &s, nil
}
