package apiary

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chnm/apiary/db"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4" // Driver for database
	"github.com/jackc/pgx/v4/pgxpool"
)

// The Config type stores configuration which is read from environment variables.
type Config struct {
	dbconn  string
	logging bool   // Whether or not to write access logs; errors/status are always logged
	address string // The address at which this will be hosted, e.g.: localhost:8090
}

// The Server type shares access to the database.
type Server struct {
	Server *http.Server
	DB     *pgxpool.Pool
	Router *mux.Router
	Config Config
}

// NewServer creates a new Server and connects to the database or fails trying.
func NewServer() *Server {
	s := Server{}

	// Read the configuration from environment variables. The `getEnv()` function
	// will provide a default.
	s.Config.dbconn = getEnv("APIARY_DB", "")
	s.Config.logging = getEnv("APIARY_LOGGING", "on") == "on"
	s.Config.address = getEnv("APIARY_INTERFACE", "0.0.0.0") + ":" + getEnv("APIARY_PORT", "8090")

	// Connect to the database then store the database in the struct.
	pool, err := db.Connect(context.TODO(), s.Config.dbconn)
	if err != nil {
		log.Fatalln("Error connecting to the database: ", err)
	}
	s.DB = pool

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

	log.Printf("Starting the server on http://%s\n", s.Config.address)

	go func() {
		if err := s.Server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	<-stop

}

// Shutdown closes the connection to the database and shutsdown the server.
func (s *Server) Shutdown() {
	log.Println("Closing the connection to the database")
	s.DB.Close()
	log.Println("Shutting down the server")
}
