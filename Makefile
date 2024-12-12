.PHONY: build clean test deps

# Binary name
BINARY_NAME=ping.exe

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Directories
CMD_DIR=cmd/ping
BUILD_DIR=build

# Build configuration
LDFLAGS=-s -w

all: deps test build

build:
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)

deps:
	$(GOGET) -v golang.org/x/net/icmp
	$(GOGET) -v golang.org/x/net/ipv4

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

install:
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o /usr/local/bin/$(BINARY_NAME) $(CMD_DIR)/main.g