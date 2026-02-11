VERSION ?= dev

.PHONY: test
test: unit-test integration-test

.PHONY: unit-test
unit-test:
	@go test ./...

.PHONY: integration-test
integration-test: build
	@test/bats/bin/bats test/*.bats

.PHONY: build
build:
	@go build -ldflags "-X github.com/juanibiapina/todo/internal/version.Version=$(VERSION)" -o dist/todo

.PHONY: install
install:
	@go install -ldflags "-X github.com/juanibiapina/todo/internal/version.Version=$(VERSION)"
