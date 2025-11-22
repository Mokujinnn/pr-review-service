
.PHONY: build run deps

all: build run

build:
	go build -o app .

run:
	docker-compose up

deps:
	go mod download
	go mod tidy
