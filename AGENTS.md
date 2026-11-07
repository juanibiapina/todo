# Project Notes

## Project Overview

`todo` is a Rust CLI and ratatui TUI for per-directory todo lists stored in one Markdown file.

- `src/main.rs`: clap command definitions and CLI output.
- `src/store.rs`: Markdown parsing, rendering and all item operations. Items are addressed by 1-based position within a directory.
- `src/tui/`: the interactive TUI: state and keys (`mod.rs`), rendering (`view.rs`), palette styles (`theme.rs`), and the text input (`input.rs`).
- `demo.tape`: records `assets/demo.gif` and `assets/help.png` with `make demo` (needs vhs). Re-record after visible TUI changes.
- `src/render.rs`: section divider formatting shared by the CLI and TUI.
- `src/config.rs`: resolves the todo file path from `TODO_FILE`.
- `test/*.bats`: CLI integration tests run against `dist/todo`.

## Build and Test

- Run check with `gob run make` (rustfmt, clippy, unit tests, bats integration tests)
- `cargo deny check bans licenses sources` checks dependency licenses and sources
- `make install` installs the binary to `~/.cargo/bin`

## Conventions

- Lints are strict (`Cargo.toml` `[lints]` and `clippy.toml`): functions stay under 100 lines and 4 levels of nesting. Extract small helpers instead of allowing the lint.
- Run `cargo fmt` before you commit.

## Release Process

- See [docs/releases.md](docs/releases.md) for the release process
