package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	creativerunrepo "github.com/woragis/ecom-op-creatives-backend/server/internal/creativerun/repository"
	creativerunsvc "github.com/woragis/ecom-op-creatives-backend/server/internal/creativerun/service"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/httpserver"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/middleware"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/migrate"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/models"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/pipeline"
	pipelinesvc "github.com/woragis/ecom-op-creatives-backend/server/internal/pipeline/service"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/postgres"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/rabbitmq"
	productrepo "github.com/woragis/ecom-op-creatives-backend/server/internal/product/repository"
	productsvc "github.com/woragis/ecom-op-creatives-backend/server/internal/product/service"
)

func main() {
	addr := envOr("HTTP_ADDR", ":8080")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		log.Fatal("RABBITMQ_URL is required")
	}

	db, err := postgres.Open(dsn)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	if skip := strings.TrimSpace(os.Getenv("SKIP_SQL_MIGRATIONS")); skip != "1" && !strings.EqualFold(skip, "true") {
		dir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR"))
		if dir == "" {
			dir = migrate.ResolveDir()
		}
		if dir != "" {
			sqlDB, err := db.DB()
			if err != nil {
				log.Fatalf("sql db: %v", err)
			}
			mctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			err = migrate.Up(mctx, sqlDB, dir)
			cancel()
			if err != nil {
				log.Fatalf("sql migrate: %v", err)
			}
			log.Printf("sql migrations applied from %s", dir)
		}
	}

	if err := db.AutoMigrate(&models.Product{}, &models.Campaign{}, &models.CreativeRun{}, &models.PipelineStep{}); err != nil {
		log.Fatalf("automigrate: %v", err)
	}

	mq, err := rabbitmq.Open(rabbitURL)
	if err != nil {
		log.Fatalf("rabbitmq: %v", err)
	}
	defer func() { _ = mq.Close() }()
	if err := mq.DeclareQueues(pipeline.AllQueues()); err != nil {
		log.Fatalf("rabbitmq declare: %v", err)
	}

	productRepository := productrepo.New(db)
	runRepository := creativerunrepo.New(db)
	pipelineService := pipelinesvc.New(mq)
	runService := creativerunsvc.New(runRepository, productRepository, pipelineService)
	productService := productsvc.New(productRepository)

	app := &httpserver.App{
		DB:       db,
		RabbitMQ: mq,
		Products: productService,
		Runs:     runService,
	}

	handler := httpserver.NewHandler(app, middleware.LoadConfigFromEnv())
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
