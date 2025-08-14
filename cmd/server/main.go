package main

import (
	"context"
	"log"
	"net/http"

	"assignment_infracloud/internal/config"
	apphttp "assignment_infracloud/internal/http"
	"assignment_infracloud/internal/service"
	"assignment_infracloud/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	shortener := service.NewInMemoryShortener(storage.NewInMemoryStore())
	srv := apphttp.NewServer(context.Background(), shortener, cfg)

	log.Printf("listening on :%s", cfg.HTTPPort)
	if err := http.ListenAndServe(":"+cfg.HTTPPort, srv); err != nil {
		log.Fatal(err)
	}
}
