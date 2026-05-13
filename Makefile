.DEFAULT_GOAL := check

.PHONY: check lint build run vet tidy init

check: lint build

lint:
	pre-commit run --all-files

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
