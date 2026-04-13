package main

import (
	"log"
	"os"

	"pharmaops/api/internal/config"
	"pharmaops/api/internal/db"
	"pharmaops/api/internal/httpserver"
)

func main() {
	cfg := config.Load()
	sqlDB, err := db.Open(cfg.DSN)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	srv := httpserver.New(cfg, sqlDB)
	if err := srv.Run(); err != nil {
		log.Printf("server: %v", err)
		os.Exit(1)
	}
}
