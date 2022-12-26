GO := go
GTEST := gotest

build:
	$(GO) build -o bin/f2bist ./cmd

test:
	$(GTEST) -v ./...
