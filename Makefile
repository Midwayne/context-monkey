GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=como
BINARY_UNIX=$(BINARY_NAME)

# Get the version from the latest git tag
VERSION := $(shell git describe --tags --abbrev=0)

# Build flags
LDFLAGS = -ldflags="-X 'como/cmd.version=$(VERSION)'"

# Default target
all: build

# Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

# Test the application
test:
	$(GOTEST) -v ./...

# Clean the binary
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe

# Cross-compilation targets
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe .

build-mac:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-mac-amd64 .

# Build all release binaries
release-build: build-linux build-windows build-mac

.PHONY: all build test clean build-linux build-windows build-mac release-build
