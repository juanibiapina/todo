# TODO

## Space shortcut should paste ticket into calling terminal
---
id: EUZ
state: refined
---
When running inside a tmux popup, the space shortcut should paste the ticket text directly into the calling tmux pane instead of just copying to clipboard. The old fzf+tmux workflow allowed this because it could send text to the parent pane. The TUI should detect if it's running in a tmux popup and use tmux send-keys to paste into the originating pane.

## Add extensive bats integration tests
---
id: USP
state: refined
---
Add comprehensive integration tests using bats, following the pattern established in juanibiapina/gob:

**Setup:**
- Add bats as a git submodule in test/bats (with bats-support and bats-assert)
- Create test/test_helper.bash with setup/teardown using BATS_TEST_TMPDIR for isolation
- Add integration-test target to Makefile (build + bats), keep unit tests separate

**Test files to create (one per command):**
- test/add.bats — add ticket, add with description, duplicate titles, empty title
- test/list.bats — list empty, list with tickets, ordering
- test/show.bats — show by ID, show by title, show nonexistent
- test/done.bats — mark done, done nonexistent, done by ID/title
- test/set_state.bats — cycle through states, invalid state
- test/set_description.bats — set, update, clear description
- test/move_up.bats — reorder tickets, move first up (no-op)
- test/move_down.bats — reorder tickets, move last down (no-op)
- test/quick_add.bats — interactive add via stdin
- test/main.bats — version flag, help, unknown command

**Test helper pattern (from gob):**
- Each test uses a fresh temp dir with its own .tickets.md
- setup() sets TODO_DIR to BATS_TEST_TMPDIR, builds binary path
- teardown() cleans up
- Helper functions for common assertions (ticket exists, ticket has state, etc.)

**Makefile changes:**
- Add integration-test target: build then run bats
- Update test target to run both unit and integration tests

## Set up release pipeline with GoReleaser and GitHub Actions
---
id: uY3
state: refined
---
Set up automated releases following the same pattern as juanibiapina/gob:

**GitHub Actions workflows:**
1. `.github/workflows/build-and-test.yaml` — CI on push to main and PRs
   - Build with `make build`
   - Run tests with `make test`
2. `.github/workflows/release.yaml` — Triggered by `v*.*.*` tags
   - Uses goreleaser/goreleaser-action@v6 with GoReleaser v2
   - Needs `contents: write` permission
   - Uses GITHUB_TOKEN and HOMEBREW_TAP_TOKEN secrets

**GoReleaser config (`.goreleaser.yaml`):**
- Build for linux/darwin on amd64/arm64 with CGO_ENABLED=0
- Inject version via ldflags: `-X github.com/juanibiapina/todo/internal/version.Version={{.Version}}`
- Archives with README.md and LICENSE.md
- SHA256 checksums
- GitHub changelog with conventional commit grouping (feat/fix/others)
- Release header/footer with download instructions
- Homebrew tap formula in juanibiapina/homebrew-taps using HOMEBREW_TAP_TOKEN

**Release docs (`docs/releases.md`):**
- Document the tag-based release process
- Semver format: `v1.2.3`, pre-releases: `v1.0.0-beta.1`
- Steps: update CHANGELOG.md, commit, tag, push tag

**Makefile update:**
- Ensure `test` target runs both unit and integration tests (once bats tests exist)

## Commands should reference tickets by ID only, not title
---
id: eEr
state: refined
---
Currently all commands accept both a 3-character ID and a title to reference tickets (via findTicket which tries ID first, then falls back to title match). Change commands to accept only the ticket ID. This simplifies the interface and avoids ambiguity when titles happen to be 3 characters long.

Commands affected: show, done, set-state, set-description, move-up, move-down.

Changes needed:
- Remove title fallback from findTicket and findTicketIndex in internal/tickets/file.go
- Update cobra Use/Long strings to say `<id>` instead of `<title|id>`
- Update help text to remove references to title lookup

## Simplify ticket states to new, refined, and done
---
id: UDK
state: refined
---
Consider removing the planned state. The current states are new, refined, planned, and done (removed from file). In practice, planned may not add value — a ticket is either new (no description), refined (well-defined problem with description), or done (completed and removed). Dropping planned simplifies state cycling, the TUI icons, and the mental model.

## Render markdown in detail panel with glamour
---
id: xoN
state: refined
---
The right-side detail panel currently displays raw text. Use charmbracelet/glamour to render the ticket description as styled markdown (headings, code blocks, lists, etc.). Glamour is the Charm ecosystem's markdown renderer and integrates naturally with bubbletea/lipgloss. It supports custom styles and width constraints, which fits the viewport panel well.
