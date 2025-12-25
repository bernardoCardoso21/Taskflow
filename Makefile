.PHONY: run test up down

up:
	docker compose up -d

down:
	docker compose down -v

run:
	go run ./cmd/api

test:
	go test ./...
