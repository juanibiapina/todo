.PHONY: test
test: lint unit-test integration-test

.PHONY: lint
lint:
	@cargo fmt --check
	@cargo clippy --all-targets --quiet -- -D warnings

.PHONY: unit-test
unit-test:
	@cargo test --quiet

.PHONY: integration-test
integration-test: build
	@test/bats/bin/bats test/*.bats

.PHONY: build
build:
	@cargo build --release --quiet
	@mkdir -p dist
	@cp target/release/todo dist/todo

.PHONY: demo
demo: build
	@PATH="$(CURDIR)/dist:$$PATH" vhs demo.tape

.PHONY: install
install:
	@cargo install --path .
