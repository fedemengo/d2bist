GO := go
GTEST := gotest

.PHONY: test

build:
	$(GO) build -o bin/d2bist .

install:
	$(GO) build -o bin/d2bist .
	$(GO) install .

test:
	$(GTEST) -v ./...
