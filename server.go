package apiary

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // Driver for database
)

// The Config type stores configuration which is read from environment variables.
type Config struct {
	dbhost  string
	dbport  string
	dbname  string
	dbuser  string
	dbpass  string
	dbssl   string // SSL mode for the database connection
	logging bool   // Whether or not to write access logs; errors/status are always logged
	address string // The address at which this will be hosted, e.g.: localhost:8090
}

// The Server type shares access to the database.
type Server struct {
	Server     *http.Server
	Database   *sql.DB
	Router     *mux.Router
	Config     Config
	Statements map[string]*sql.Stmt
}

// NewServer creates a new Server and connects to the database or fails trying.
func NewServer() *Server {
	s := Server{}

	// Read the configuration from environment variables. The `getEnv()` function
	// will provide a default.
	s.Config.dbhost = getEnv("DATAAPI_DBHOST", "localhost")
	s.Config.dbport = getEnv("DATAAPI_DBPORT", "5432")
	s.Config.dbname = getEnv("DATAAPI_DBNAME", "")
	s.Config.dbuser = getEnv("DATAAPI_DBUSER", "")
	s.Config.dbpass = getEnv("DATAAPI_DBPASS", "")
	s.Config.dbssl = getEnv("DATAAPI_SSL", "require")
	s.Config.logging = getEnv("DATAAPI_LOGGING", "on") == "on"
	s.Config.address = getEnv("DATAAPI_INTERFACE", "0.0.0.0") + ":" + getEnv("DATAAPI_PORT", "8090")

	// Connect to the database then store the database in the struct.
	constr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		s.Config.dbhost, s.Config.dbport, s.Config.dbname, s.Config.dbuser,
		s.Config.dbpass, s.Config.dbssl)
	db, err := sql.Open("postgres", constr)
	if err != nil {
		log.Fatalln(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalln(err)
	}
	s.Database = db

	// Create an empty map to store prepared statements
	s.Statements = make(map[string]*sql.Stmt)

	// Create the router, store it in the struct, initialize the routes, and
	// register the middleware.
	router := mux.NewRouter()
	s.Router = router
	s.Routes()
	s.Middleware()

	s.Server = &http.Server{
		Addr:         s.Config.address,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      s.Router,
	}

	return &s
}

// Run starts the API server.
func (s *Server) Run() {
	defer s.Shutdown() // Make sure we shutdown.

	// Run the server in a go routine, using a blocking channel to listen for interrupts.
	// Stop gracefully for SIGTERM and SIGINT.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	log.Printf("Starting the server on http://%s.\n", s.Config.address)

	go func() {
		if err := s.Server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	<-stop

}

// Shutdown closes the connection to the database and shutsdown the server.
func (s *Server) Shutdown() {
	// Close any prepared statements
	for _, v := range s.Statements {
		err := v.Close()
		if err != nil {
			log.Println(err)
		}
	}
	log.Println("Closing the connection to the database.")
	err := s.Database.Close()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Shutting down the server.")
}
