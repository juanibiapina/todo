# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Render markdown in TUI detail panel using glamour
- Comprehensive bats integration tests
- Space shortcut in TUI to copy ticket ID to clipboard
- `todo add` flags: `-d/--description`, `-t/--type` (default `task`), `-p/--priority` (default `2`), `-a/--assignee` (default `git user.name`), `--external-ref`, `--parent` (validates existence), `--design`, `--acceptance`, `--tags` (comma-separated)
- Default title "Untitled" when `todo add` is called with no arguments
- Parent ticket validation: `--parent` must reference an existing ticket ID
- `todo status <id> <status>` — set ticket status directly (valid: `open`, `in_progress`, `closed`)
- `todo start <id>` — shortcut to set status to `in_progress`
- `todo close <id>` — shortcut to set status to `closed`
- `todo reopen <id>` — shortcut to set status to `open`
- Partial ID matching: all commands accepting a ticket ID now support substring matching (exact match takes precedence; ambiguous matches produce an error)

### Changed

- **Breaking:** Tickets now stored as individual files in `docs/tickets/` directory
  - Each ticket is a separate file named `<id>.md`
  - File format uses YAML frontmatter (`---` delimited) followed by `# Title` heading and description
  - Enables better git diffs and easier manual editing
- `todo done` now sets `status: closed` instead of deleting the ticket file. Closed tickets are preserved on disk but hidden from `list` and TUI.
- All commands now reference tickets by ID only, not by title
  - Affects: `show`, `done`, `set-description`
  - Title fallback removed from internal lookup functions

### Removed

- Removed `move-up` and `move-down` commands (ordering no longer supported)
- Removed `K`/`J` reorder keybindings from TUI
- Removed ticket states entirely (no more `new`/`refined` distinction)
  - Removed `set-state` command
  - Removed state icons from list and TUI output
  - Removed `s`/`S` keybindings from TUI
  - Removed `state:` field from file format
  - Simpler output: `list` now shows just `ID Title`

### Fixed

- TUI no longer creates tickets directory on startup; directory is only created when adding a ticket

## [0.2.0] - 2026-02-11

### Changed

- Renamed tickets file from `.tickets.md` to `TODO.md`

### Added

- `todo tui` - Full-screen bubbletea TUI with split-panel layout (ticket list + detail view)
  - Navigate, add, delete, cycle state, reorder tickets interactively
  - Help modal with `?`
  - Status bar with contextual keybindings
  - `esc` key to quit
- `todo quick-add` - Interactive prompt for adding tickets (designed for tmux popup)
- `todo move-up` / `todo move-down` - Reorder tickets via CLI
- State cycling: `NextState`, `PrevState`, `CycleState`, `CycleStateBack`
- State icons (nerd font): ○ new, ◐ refined, ● planned

## [0.1.0] - 2026-02-11

### Added

- `todo add` - Create tickets with optional description (via argument or stdin)
- `todo list` - List all tickets with ID, state, and title
- `todo show` - Show full ticket details by ID or title
- `todo done` - Remove a ticket (mark complete)
- `todo set-state` - Change ticket state (new, refined, planned)
- `todo set-description` - Set/replace description (via argument or stdin)
- Stdin support for descriptions: heredocs and pipes work with backticks, code blocks, and special characters
- Tickets stored in `.tickets.md` in the current directory
- 3-character base62 ticket IDs
