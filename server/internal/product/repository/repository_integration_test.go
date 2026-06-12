//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/woragis/ecom-op-creatives-backend/server/internal/models"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/postgres"
	"gorm.io/gorm"
)

func TestProductRepositoryCreateAndGet(t *testing.T) {
	dsn := integrationDSN(t)
	db, err := postgres.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sqlDB.Close() }()

	resetProducts(t, db)

	repo := New(db)
	product := &models.Product{Name: "Cable Organizer"}
	if err := repo.Create(context.Background(), product); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetByID(context.Background(), product.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Cable Organizer" {
		t.Fatalf("name = %q", got.Name)
	}
}

func resetProducts(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec("TRUNCATE products CASCADE").Error; err != nil {
		t.Fatal(err)
	}
}
