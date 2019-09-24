package main

import (
	"log"
	"net/http"
)

func main() {
	// Connect to the database, but make sure it gets closed
	s, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}
	defer s.Database.Close()

	// Setup the routes
	s.routes()

	log.Println("Starting the server ...")
	log.Fatalln(http.ListenAndServe(":8080", s.Router))

}
