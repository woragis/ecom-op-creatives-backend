.PHONY: test dev migrate worker-stub worker-pipeline

test:
	cd server && go test ./...

dev:
	cd server && go run ./cmd/server

migrate:
	cd server && go run ./cmd/migrate

worker-stub:
	cd server && go run ./cmd/worker-stub

worker-pipeline:
	cd server && go run ./cmd/worker-pipeline
