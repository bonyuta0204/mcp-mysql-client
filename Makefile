DIST := bin
BIN  := $(DIST)/mcp-mysql-client
GO   := go
SRCS = $(shell find . -type f -name "*.go")

.PHONY: run build test test-unit test-e2e docker-up docker-down



build: $(BIN)

$(BIN): $(DIST) $(SRCS)
	$(GO) build -o $@

$(DIST):
	@mkdir -p $(DIST)

run: build
	./$(BIN)

# Run all tests
test: test-unit test-e2e

# Run unit tests
test-unit:
	$(GO) test -v ./pkg/...

# Run end-to-end tests (requires MySQL)
test-e2e: build
	$(GO) test -v ./e2e/...

# Start MySQL container for tests
docker-up:
	docker-compose up -d
	@echo "Waiting for MySQL to be ready..."
	@sleep 10

# Stop and remove MySQL container
docker-down:
	docker-compose down