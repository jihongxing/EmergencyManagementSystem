package main

import (
	"context"
	"log"
	"os"
	"time"

	"emergency-management/backend/internal/database"
)

func main() {
	db, err := database.Open(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := database.Migrate(ctx, db); err != nil {
		log.Fatal("migration failed; check connectivity and migration configuration")
	}
	log.Print("migration baseline applied")
}
