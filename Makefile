.PHONY: run build test test-coverage lint migrate-up migrate-down seed docker-build

APP_NAME := hookrelay
BINARY := bin/$(APP_NAME)
MAIN := cmd/server/main.go

run:
	go run $(MAIN)

build:
	CGO_ENABLED=0 go build -o $(BINARY) $(MAIN)

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

migrate-up:
	go run $(MAIN) -migrate-up

migrate-down:
	go run $(MAIN) -migrate-down

seed:
	go run seed/main.go

docker-build:
	docker build -t $(APP_NAME) .
