.PHONY: all build up
all: build

build:
	-mkdir -p ./build/
	CGO_ENABLED=0 go build -v -o ./build/ ./...

	docker-compose build

up:
	docker-compose up