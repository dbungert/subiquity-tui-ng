.PHONY: run build vet tidy

run:
	go run .

build:
	go build ./...

vet:
	go vet ./...

tidy:
	go mod tidy
