.DEFAULT_GOAL := check

.PHONY: check lint test coverage build run vet tidy init

check: lint test build

lint:
	pre-commit run --all-files

test:
	go test -cover ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

build:
	go build ./...

run:
	go run .

vet:
	go vet ./...

tidy:
	go mod tidy

init:
	pre-commit install
