.PHONY: run build docker-up docker-down generate migrate-up migrate-down test test-integration lint

# --------------------------------------------------
# Development
# --------------------------------------------------

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

# --------------------------------------------------
# Docker
# --------------------------------------------------

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-build:
	docker-compose build

# --------------------------------------------------
# Code generation
# --------------------------------------------------

generate:
	@echo "Running sqlc generate..."
	sqlc generate
	@echo "Running gqlgen generate..."
	go run github.com/99designs/gqlgen generate

# --------------------------------------------------
# Database migrations
# --------------------------------------------------

MIGRATE_DNS ?="postgres://bloguser:blogpass@localhost:5432/blogdb?sslmode=disable"

migrate-up:
	migrate -database ${MIGRATE_DNS} -path migrations up

migrate-down:
	migrate -database ${MIGRATE_DNS} -path migrations down

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# --------------------------------------------------
# Testing
# --------------------------------------------------

test:
	go test ./... -race -count=1

test-integration:
	go test ./... -race -tags=integration -count=1

test-coverage:
	go test ./... -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# --------------------------------------------------
# Linting
# --------------------------------------------------

lint:
	golangci-lint run ./...