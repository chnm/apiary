package dataapi

import (
	"database/sql"
	"fmt"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // Driver for database
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
	dbhost := getEnv("DATAAPI_DBHOST", "localhost")
	dbport := getEnv("DATAAPI_DBPORT", "5432")
	dbname := getEnv("DATAAPI_DBNAME", "dataapi")
	dbuser := getEnv("DATAAPI_DBUSER", "dataapi")
	dbpass := getEnv("DATAAPI_DBPASS", "")
	dbsslm := getEnv("DATAAPI_SSL", "disable")

	constr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		dbhost, dbport, dbname, dbuser, dbpass, dbsslm)
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
