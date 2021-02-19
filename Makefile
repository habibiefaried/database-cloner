all: build

build:
	go build -ldflags "-linkmode external -extldflags -static" main.go
	chmod +x main

format:
	go fmt ./...