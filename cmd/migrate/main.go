package main

import (
	"context"
	"log"
	"path/filepath"

	"skeleton-go/internal/migration"
	"skeleton-go/pkg/config"
	"skeleton-go/pkg/database"
)

func main() {
	cfg := config.Load()

	db, err := database.NewMySQL(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	runner := migration.NewRunner(db, filepath.Join("migrations"))
	if err := runner.Run(context.Background()); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migrations applied successfully")
}
