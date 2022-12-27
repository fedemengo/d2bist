GO := go
GTEST := gotest

.PHONY: test

build:
	$(GO) build -o bin/f2bist ./cmd

test:
	$(GTEST) -v ./...
