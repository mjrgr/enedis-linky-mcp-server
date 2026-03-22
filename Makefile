BINARY     := enedis-linky-mcp-server
BUILD_DIR  := ./bin
CMD        := ./cmd/server
LDFLAGS    := -s -w

.PHONY: all build run test lint clean docker-build docker-run help

all: build

## build: Compile the binary into ./bin/
build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) $(CMD)
	@echo "Built $(BUILD_DIR)/$(BINARY)"

## run: Run the server (stdio transport by default). Set MCP_TRANSPORT=sse for HTTP.
run:
	@[ -f .env ] && export $$(grep -v '^#' .env | xargs) ; \
	go run $(CMD)

## test: Run all tests with race detector
test:
	go test -race -count=1 ./...

## test-cover: Run tests and open coverage report
test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

## lint: Run golangci-lint (install: https://golangci-lint.run/usage/install/)
lint:
	golangci-lint run ./...

## tidy: Tidy go.mod and go.sum
tidy:
	go mod tidy

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR) coverage.out

## docker-build: Build the Docker image
docker-build:
	docker build -t $(BINARY):latest .

## docker-run: Run the Docker image (requires CONSO_API_TOKEN env var)
docker-run:
	docker run --rm \
		-e CONSO_API_TOKEN="$(CONSO_API_TOKEN)" \
		-e MCP_TRANSPORT=sse \
		-p 8080:8080 \
		$(BINARY):latest

## help: Show this help message
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
