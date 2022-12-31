GO := go
GTEST := gotest

.PHONY: test

build:
	$(GO) build -o bin/d2bist .

test:
	$(GTEST) -v ./...
