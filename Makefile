VERSION ?= dev

.PHONY: test
test:
	@go test ./...

.PHONY: build
build:
	@go build -ldflags "-X github.com/juanibiapina/todo/internal/version.Version=$(VERSION)" -o dist/todo

.PHONY: install
install:
	@go install -ldflags "-X github.com/juanibiapina/todo/internal/version.Version=$(VERSION)"
