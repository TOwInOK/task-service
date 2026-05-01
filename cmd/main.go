package main

import (
	"fmt"
	"log"
	"net/http"

	"task-service/internal/config"
	"task-service/internal/storage"

	httphandler "task-service/internal/http"
)

func main() {
	if err := config.Load("config.json"); err != nil {
		log.Fatalf("config load: %v", err)
	}
	cfg := config.Get()
	log.Printf("Loaded config: %+v", cfg)

	// Initialize storage actor
	actor, err := storage.New("./data")
	if err != nil {
		log.Fatalf("storage init: %v", err)
	}
	defer actor.Stop()

	// Initialize HTTP server
	srv := httphandler.NewServer(actor)

	log.Printf("Starting server on %s", cfg.Port)
	if err := http.ListenAndServe(cfg.Port, srv.Router); err != nil {
		log.Fatalf("server: %v", err)
	}
	fmt.Println("done")
}
