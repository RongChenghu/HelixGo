.PHONY: dev ping-bot build test tidy

dev:
	go run ./cmd/helix-bot

# Run ping-bot example: /ping -> pong (minimal acceptance).
ping-bot:
	go run ./examples/ping-bot

build:
	go build -o bin/helix-bot ./cmd/helix-bot

test:
	go test ./...

tidy:
	go mod tidy
