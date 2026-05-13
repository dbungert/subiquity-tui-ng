.DEFAULT_GOAL := check

.PHONY: check lint test coverage build run init

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

init:
	pre-commit install
