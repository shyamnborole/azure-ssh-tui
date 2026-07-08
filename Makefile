.PHONY: all build run test lint clean

APP_NAME = azure-ssh-tui
GO_FILES = $(shell find . -type f -name '*.go')

all: build

build:
	go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

run:
	go run cmd/$(APP_NAME)/main.go

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
