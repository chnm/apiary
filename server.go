package apiary

import (
	"context"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/chnm/apiary/db"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4" // Driver for database
	"github.com/jackc/pgx/v4/pgxpool"

	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
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
	Cache  *cache.Client
}

// NewServer creates a new Server and connects to the database or fails trying.
func NewServer(ctx context.Context) *Server {
	s := Server{}

	// Read the configuration from environment variables. The `getEnv()` function
	// will provide a default.
	s.Config.dbconn = getEnv("APIARY_DB_LOCAL", "")
	s.Config.logging = getEnv("APIARY_LOGGING", "on") == "on"
	s.Config.address = getEnv("APIARY_INTERFACE", "0.0.0.0") + ":" + getEnv("APIARY_PORT", "8090")

	// Connect to the database then store the database in the struct.
	log.Println("connecting to the database")
	dbTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	pool, err := db.Connect(dbTimeout, s.Config.dbconn)
	if err != nil {
		log.Fatalln("error connecting to the database:", err)
	}
	s.DB = pool

	// Set up the in-memory cache
	memcached, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(int(math.Pow(1000, 2))), // 1000^3 = gigabyte
	)
	if err != nil {
		log.Fatal("error setting up memory cache:", err)
	}

	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(1*time.Hour),
		cache.ClientWithRefreshKey("nocache"),
	)
	if err != nil {
		log.Fatal("error setting up memory cache:", err)
	}
	s.Cache = cacheClient

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
func (s *Server) Run() error {
	log.Printf("starting the server on http://%s\n", s.Config.address)
	err := s.Server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Shutdown closes the connection to the database and shutsdown the server.
func (s *Server) Shutdown() {
	log.Println("closing the connection to the database")
	s.DB.Close()
	log.Println("shutting down the web server")
	err := s.Server.Shutdown(context.TODO())
	if err != nil {
		log.Println("error shutting down web server:", err)
	}
}
