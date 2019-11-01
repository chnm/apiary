package relecapi

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // Driver for database
)

// The Server type shares access to the database.
type Server struct {
	Database *sql.DB
	Router   *mux.Router
}

// NewServer creates a new Server and connects to the database or fails trying.
func NewServer() *Server {
	s := Server{}

	// Connect to the database. Read the configuration from environment variables,
	// connect to the database, and then store the database in the struct.
	dbhost := getEnv("RELECAPI_DBHOST", "localhost")
	dbport := getEnv("RELECAPI_DBPORT", "5432")
	dbname := getEnv("RELECAPI_DBNAME", "")
	dbuser := getEnv("RELECAPI_DBUSER", "")
	dbpass := getEnv("RELECAPI_DBPASS", "")
	dbsslm := getEnv("RELECAPI_SSL", "")

	constr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		dbhost, dbport, dbname, dbuser, dbpass, dbsslm)
	db, err := sql.Open("postgres", constr)
	if err != nil {
		log.Fatalln(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalln(err)
	}
	s.Database = db

	// Create the router, store it in the struct, initialize the routes, and
	// register the middleware.
	router := mux.NewRouter()
	s.Router = router
	s.Routes()
	s.Middleware()

	return &s
}

// Run starts the API server.
func (s *Server) Run() {
	defer s.Shutdown() // Make sure we shutdown.

	// Run the server in a go routine, using a blocking channel to listen for interrupts.
	// Stop gracefully for SIGTERM and SIGINT.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	port := ":" + getEnv("RELECAPI_PORT", "8090")

	log.Printf("Starting the server on http://localhost%s.\n", port)
	go func() {
		err := http.ListenAndServe(port, s.Router)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	<-stop

}

// Shutdown closes the connection to the database and shutsdown the server.
func (s *Server) Shutdown() {
	log.Println("Closing the connection to the database.")
	err := s.Database.Close()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Shutting down the server.")
}
