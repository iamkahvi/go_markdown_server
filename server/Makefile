SHELL := /bin/bash

.PHONY: run build test lint

run:
	DEV=1 go run ./cmd/server

build:
	go build ./cmd/server

test:
	go test ./...

lint:
	golangci-lint run
