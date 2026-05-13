.DEFAULT_GOAL := check
COVERAGE_FILE := .coverage.out

.PHONY: check lint test coverage build run init

check: lint test build

lint:
	pre-commit run --all-files

test:
	go test -cover ./...

coverage:
	go test -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -func=$(COVERAGE_FILE)

build:
	go build -o subiquity-client ./cmd/subiquity-client

run:
	go run ./cmd/subiquity-client

init:
	pre-commit install
