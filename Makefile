.PHONY: test dev migrate worker-stub

test:
	cd server && go test ./...

dev:
	cd server && go run ./cmd/server

migrate:
	cd server && go run ./cmd/migrate

worker-stub:
	cd server && go run ./cmd/worker-stub
