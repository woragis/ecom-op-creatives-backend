package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/migrate"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/postgres"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	dir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR"))
	if dir == "" {
		dir = migrate.ResolveDir()
	}
	if dir == "" {
		log.Fatal("MIGRATIONS_DIR is not set and no migrations/ directory was found")
	}

	db, err := postgres.Open(dsn)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("sql db: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := migrate.Up(ctx, sqlDB, dir); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Printf("migrate: OK (dir=%s)", dir)
}
