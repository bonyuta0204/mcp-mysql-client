DIST := bin
BIN  := $(DIST)/mcp-mysql-client
GO   := go
SRCS = $(shell find . -type f -name *.go)

.PHONY: run build



build: $(BIN)

$(BIN): $(DIST) $(SRCS)
	$(GO) build -o $@

$(DIST):
	@mkdir -p $(DIST)

run: build
	./$(BIN)
