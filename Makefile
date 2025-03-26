DIST := bin
BIN  := $(DIST)/mcp-mysql-client
GO   := go
SRCS = $(shell find . -type f -name *.go)

.PHONY: run build test test-unit test-integration docker-up docker-down



build: $(BIN)

$(BIN): $(DIST) $(SRCS)
	$(GO) build -o $@

$(DIST):
	@mkdir -p $(DIST)

run: build
	./$(BIN)

test: test-unit test-integration

test-unit:
	$(GO) test -v ./pkg/...

# Run integration tests (requires MySQL)
test-integration:
	$(GO) test -v ./pkg/integration/...

# Start MySQL container for integration tests
docker-up:
	docker-compose up -d
	@echo "Waiting for MySQL to be ready..."
	@sleep 10

# Stop and remove MySQL container
docker-down:
	docker-compose down
