package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	creativerunrepo "github.com/woragis/ecom-op-creatives-backend/server/internal/creativerun/repository"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/config"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/migrate"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/models"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/app"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/pipeline"
	pipelinesvc "github.com/woragis/ecom-op-creatives-backend/server/internal/pipeline/service"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/postgres"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/rabbitmq"
	productrepo "github.com/woragis/ecom-op-creatives-backend/server/internal/product/repository"
)

func main() {
	cfg := config.Load()
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
		dir := migrate.ResolveDir()
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

	runRepo := creativerunrepo.New(db)
	productRepo := productrepo.New(db)
	pipelineSvc := pipelinesvc.New(mq)
	exec, err := app.NewExecutor(cfg, runRepo, productRepo, pipelineSvc)
	if err != nil {
		log.Fatalf("executor: %v", err)
	}

	for _, queue := range pipeline.AllQueues() {
		q := queue
		if err := mq.Consume(q, func(d amqp.Delivery) error {
			return handleJob(context.Background(), exec, d.Body)
		}); err != nil {
			log.Fatalf("consume %s: %v", q, err)
		}
		log.Printf("worker-pipeline consuming %s", q)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Print("worker-pipeline shutting down")
}

func handleJob(ctx context.Context, exec interface {
	ProcessStep(ctx context.Context, stepID uuid.UUID) error
}, body []byte) error {
	var msg rabbitmq.JobMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return err
	}
	stepID, err := uuid.Parse(msg.StepID)
	if err != nil {
		return err
	}
	log.Printf("processing step %s (%s)", msg.StepID, msg.StepType)
	return exec.ProcessStep(ctx, stepID)
}
