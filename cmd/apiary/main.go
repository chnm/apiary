package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	apiary "github.com/chnm/apiary"
)

func main() {

	var server *apiary.Server
	// Create a context and listen for signals to gracefully shutdown the application
	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// Clean up function that will be called at program end no matter what
	defer func() {
		signal.Stop(quit)
		cancel()
	}()

	// Listen for shutdown signals in a go-routine and cancel context then
	go func() {
		select {
		case <-quit:
			log.Println("shutdown signal received, so quitting Apiary")
			cancel()
			server.Shutdown()
		case <-ctx.Done():
		}
	}()

	server = apiary.NewServer(ctx)
	err := server.Run()
	if err != nil {
		log.Fatal("error running the server:", err)
	}

}
