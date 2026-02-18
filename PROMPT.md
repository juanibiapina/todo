# Todo Feature Parity with wedow/ticket

## Goal

Bring all features from wedow/ticket into juanibiapina/todo, keeping todo's unique features (TUI, quick-add, set-description, color output). Add comprehensive bats integration tests matching wedow/ticket's BDD coverage.

## Context

**Current state:** Go CLI tool (`github.com/juanibiapina/todo`) using cobra for commands, with bubbletea TUI. Has `add`, `done`, `list`, `show`, `set-description`, `quick-add`, and `tui` commands. Tickets stored as markdown in `docs/tickets/` with format `# Title\n---\nid: XXX\n---\nDescription`. Files named `<id>-<slug>.md` with 3-char random base62 IDs.

**Technologies:** Go 1.25, cobra CLI, bubbletea/lipgloss TUI, bats integration tests, go test unit tests. Build via Makefile (`make test` runs both).

**Key files:**
- `internal/tickets/ticket.go` — Ticket struct (Title, ID, Description)
- `internal/tickets/file.go` — File I/O, List, Add, Show, Done, SetDescription
- `internal/tickets/id.go` — 3-char base62 ID generation
- `cmd/*.go` — cobra commands
- `test/*.bats` — integration tests

**Design decisions:**
- `done` → sets status=closed (no file deletion)
- File format → standard YAML frontmatter first (`--- YAML --- # Title`)
- File naming → `<id>.md` only (no slug)
- ID format → 3-char random base62 IDs (current format)
- Plugin system → skip
- Query → Go-native JSON with built-in filtering
- Tickets directory → keep `docs/tickets/`

## Requirements

- Maintain backward compatibility with existing TUI, quick-add, set-description, and color output features
- All new features must have bats integration tests
- Existing tests must be updated to match new formats
- File format changes must be applied consistently across all commands
- ID resolution must support partial matching across all commands

## Steps

- [x] Update Ticket struct to add new fields (Status, Deps, Links, Created, Type, Priority, Assignee, ExternalRef, Parent, Tags) and change FullString() to write YAML-frontmatter-first format (`--- YAML --- # Title\nDescription`). Add unit tests for the new format (iteration 1)
- [x] Keep 3-char random base62 ID generation (current format). No changes needed to id.go. Verify existing unit tests cover ID uniqueness and format (iteration 2)
- [x] Change file naming from `<id>-<slug>.md` to `<id>.md`. Remove slugify(), update ticketFileName/ticketFilePath, update findTicketFile for exact match, update parseFile/writeFile for YAML-first frontmatter format. Update file_test.go unit tests (iteration 3)
- [x] Update all commands (add, done, show, list, set-description, format) to work with the new file format, naming, and ID generation. Update all existing bats tests to match new output format, ID patterns, and done behavior (status=closed instead of delete) (iteration 4)
- [x] Add creation flags to add command: `-d/--description`, `-t/--type` (bug/feature/task/epic/chore, default task), `-p/--priority` (0-4, default 2), `-a/--assignee` (default git user.name), `--external-ref`, `--parent` (validate exists), `--design`, `--acceptance`, `--tags` (comma-separated). Default title to "Untitled". Add bats tests for each flag and default values (iteration 5)
- [x] Add status management commands: `status <id> <status>` (validate open|in_progress|closed), `start <id>`, `close <id>`, `reopen <id>` shortcuts. Change `done` to set status=closed instead of deleting. Add bats tests for each command, invalid status, and non-existent ticket errors (iteration 6)
- [x] Enhance findTicketFile with partial ID resolution: exact match first, then glob `*<id>*.md`, error on ambiguous matches. Apply to all commands that take an ID. Add bats tests for exact/prefix/suffix/substring matches, ambiguous errors, and exact precedence (iteration 7)
- [x] Add dep/undep commands for dependency management: `dep <id> <dep-id>` (idempotent, validates both exist), `undep <id> <dep-id>`. Store deps as YAML array. Add bats tests for add/remove, idempotency, and validation errors (iteration 8)
- [ ] Add dep tree command with box-drawing output (`├── `, `└── `, `│   `), `--full` flag for no dedup, `[status]` and title per node, sorted by subtree depth then ID, cycle-safe. Add bats tests for tree format, sorting, cycles, and full mode
- [ ] Add dep cycle command: DFS-based cycle detection on open tickets, output normalized cycles with member details. Add bats tests
- [ ] Add link/unlink commands for bidirectional linking: `link <id> <id> [id...]` updates all involved files (idempotent), `unlink <id> <target-id>` removes from both. Add bats tests for bidirectional links, 3+ tickets, idempotency, and unlink
- [ ] Enhance list command with `--status`, `-a/--assignee`, `-T/--tag` filters. Show deps in output: `id [status] - Title <- [dep1, dep2]`. Empty list returns nothing instead of "No tickets". Add bats tests for all filter combinations
- [ ] Add ready command: show open/in_progress tickets with all deps closed or no deps, sorted by priority then ID, format `id [P2][open] - Title`, support assignee/tag filters. Add bats tests
- [ ] Add blocked command: show open/in_progress tickets with unclosed deps, show only unclosed blockers in output, support assignee/tag filters. Add bats tests
- [ ] Add closed command: show recently closed tickets sorted by mtime, `--limit=N` (default 20), support assignee/tag filters. Add bats tests
- [ ] Enhance show command to compute relationships by loading all tickets: append Blockers (unclosed deps), Blocking (reverse deps), Children (parent matches), and Linked sections. Enhance parent line with title. Support `TODO_PAGER` env var. Add bats tests for each computed section
- [ ] Add add-note command: `add-note <id> [text]` appends `## Notes` section if missing, then `**<timestamp>**\n\n<text>`, support stdin pipe. Add bats tests
- [ ] Add edit command: `edit <id>` opens ticket in `$EDITOR` (default vi), print file path if non-TTY. Add bats tests
- [ ] Add query command: output all tickets as JSONL with all frontmatter fields, support `--status`, `--type`, `--assignee`, `--tag` filters. Go-native implementation. Add bats tests for JSONL validity, field presence, filtering, and empty output

## Learnings

- Used a separate `frontmatter` helper struct to exclude Title and Description from YAML marshaling — they render in the markdown body instead
- `omitempty` on all YAML fields except `id` keeps output minimal; priority=0 is omitted which is acceptable since step 5 sets default priority=2
- `gopkg.in/yaml.v3` added as dependency for proper YAML serialization
- `generateID()` and `generateUniqueID()` are unexported — tests must be in `package tickets` (same package) to access them directly
- Only prior ID test coverage was a `len(ticket.ID) != 3` check in `file_test.go` — dedicated `id_test.go` now provides comprehensive coverage
- `findTicketFile()` simplified from glob to exact `os.Stat()` — more efficient since filename is deterministic from ID (`<id>.md`)
- `SetDescription()` no longer needs file rename since filename doesn't depend on title — just overwrites in place
- Search for `\n---\n` as closing frontmatter delimiter correctly handles descriptions containing `---` on their own line
- `Done()` sets `t.Status = "closed"` then calls `writeFile()` — preserves ticket data on disk instead of deleting
- `List()` remains a pure data function returning all tickets (including closed); filtering happens in `cmd/list.go` and `internal/tui/tui.go` — keeps the data layer flexible for future `--status` filtering and `closed` command
- Most commands (add, show, set-description, format) needed no changes for step 4 — they already worked with the new YAML frontmatter format from iterations 1-3
- Refactored `Add()` to accept `*Ticket` struct instead of individual parameters — cleaner API that avoids parameter explosion as fields grow
- Parent validation done in `Add()` data layer (not cobra command) — ensures consistency regardless of entry point (CLI, TUI, tests)
- Description priority order: `-d` flag > positional arg > stdin — most explicit input wins
- Default assignee uses `git config user.name`; detected via `cmd.Flags().Changed("assignee")` to only apply when flag not explicitly set
- `SetStatus()` centralizes validation and status mutation at the library level using a `validStatuses` map — all 4 commands (status, start, close, reopen) delegate to it for consistent behavior
- Shortcut commands (`start`, `close`, `reopen`) call `SetStatus()` with hardcoded status values rather than duplicating validation logic
- `done` command left unchanged and coexists with `close` — both set status=closed, different output messages
- `findTicketFile()` uses substring matching (`strings.Contains`) not just prefix — allows prefix, suffix, and interior substring matches on ticket IDs
- No signature change to `findTicketFile` means all callers (show, done, close, start, reopen, status, set-description, parent validation) automatically get partial ID support without any code changes
- Ambiguous match error uses `sort.Strings(ids)` for deterministic, testable error messages
- `AddDep()` and `RemoveDep()` resolve partial IDs via `findTicketFile` and store the full resolved ID in the deps array — ensures consistent references even when users provide partial IDs
- Both `AddDep` and `RemoveDep` are idempotent — `AddDep` checks existing deps before appending, `RemoveDep` filters without erroring if dep not present
- Both ticket and dependency ticket must exist (validated via `findTicketFile`) before any modification — prevents dangling references in the deps array

## History

### Iteration 1: Ticket struct fields and YAML frontmatter FullString()
- **Branch**: ralph/ticket-yaml-frontmatter
- **PR**: #2 (merged)
- **Summary**: Added 10 new fields to Ticket struct (Status, Type, Priority, Assignee, Created, Parent, ExternalRef, Deps, Links, Tags). Created `frontmatter` helper struct with YAML tags. Rewrote `FullString()` to output `---\nYAML\n---\n# Title\nDescription` format using `gopkg.in/yaml.v3`. Added 12 unit tests in `ticket_test.go`. `String()` method unchanged.

### Iteration 2: ID generation test coverage
- **Branch**: ralph/id-generation-tests
- **PR**: #3 (merged)
- **Summary**: Created `internal/tickets/id_test.go` with 6 test functions covering `generateID()` (length, base62 character set, randomness) and `generateUniqueID()` (empty map, collision avoidance, high-pressure with 1000 pre-populated IDs). No changes to `id.go` — verification only.

### Iteration 3: ID-only filenames and YAML frontmatter parsing
- **Branch**: ralph/id-only-filenames-and-parse-yaml
- **PR**: #4 (merged)
- **Summary**: Simplified file naming from `<id>-<slug>.md` to `<id>.md`. Removed `slugify()` and `regexp`/`bufio` imports. Simplified `findTicketFile()` to exact `os.Stat()` check. Rewrote `parseFile()` to read YAML-frontmatter-first format using `gopkg.in/yaml.v3` unmarshal into `frontmatter` struct, populating all 13 Ticket fields. Simplified `SetDescription()` to overwrite in place (no rename). Updated `file_test.go`: removed `TestSlugify`, rewrote `TestTicketFileName`/`TestFileFormat`/`TestDone`, added `TestParseFileRoundTripAllFields` and `TestParseFileDescriptionWithDashes`. Updated `README.md` and `CHANGELOG.md`. All 32 unit tests and 29 bats tests pass.

### Iteration 8: Dep/undep commands for dependency management
- **Branch**: ralph/dep-undep-commands
- **PR**: #9 (merged)
- **Summary**: Added `AddDep(dir, id, depID string)` and `RemoveDep(dir, id, depID string)` to `internal/tickets/file.go` — both validate both IDs exist via `findTicketFile`, store full resolved IDs, and are idempotent. Created `cmd/dep.go` and `cmd/undep.go` cobra commands with `ExactArgs(2)`. Added 7 unit tests (`TestAddDep`, `TestAddDepIdempotent`, `TestAddDepTicketNotFound`, `TestAddDepDepNotFound`, `TestRemoveDep`, `TestRemoveDepNotPresent`, `TestRemoveDepTicketNotFound`). Created `test/dep.bats` (8 tests) and `test/undep.bats` (5 tests). Updated README.md and CHANGELOG.md. All 50 unit tests and 93 bats tests pass.

### Iteration 7: Partial ID matching
- **Branch**: ralph/partial-id-matching
- **PR**: #8 (merged)
- **Summary**: Enhanced `findTicketFile()` with partial ID resolution — exact match first via `os.Stat()`, then `os.ReadDir()` scanning `.md` files for substring matches using `strings.Contains()`. Added `sort` and `strings` imports. Ambiguous matches (2+) produce sorted error message. Added 6 unit tests (`TestFindTicketFilePartialPrefix`, `Suffix`, `Substring`, `Ambiguous`, `ExactPrecedence`, `NotFound`). Created `test/partial_id.bats` with 11 integration tests covering show/done/status/start/close/reopen/set-description with partial IDs, not-found error, and exact match precedence. Updated README.md and CHANGELOG.md. All 43 unit tests and 80 bats tests pass.

### Iteration 6: Status management commands
- **Branch**: ralph/status-management-commands
- **PR**: #7 (merged)
- **Summary**: Added `SetStatus(dir, id, status string)` function with `validStatuses` map for validation. Created 4 new commands: `status <id> <status>`, `start <id>` (sets in_progress), `close <id>` (sets closed), `reopen <id>` (sets open). Added 3 unit tests (`TestSetStatus`, `TestSetStatusInvalid`, `TestSetStatusNotFound`). Created 4 bats test files with 18 integration tests (`test/status.bats` 7, `test/start.bats` 3, `test/close.bats` 4, `test/reopen.bats` 4). Updated README.md and CHANGELOG.md. All 37 unit tests and 69 bats tests pass.

### Iteration 5: Add command creation flags
- **Branch**: ralph/add-command-flags
- **PR**: #6 (merged)
- **Summary**: Added `Design` and `Acceptance` fields to `Ticket` and `frontmatter` structs. Refactored `Add(dir, title, description)` to `Add(dir string, t *Ticket)` for cleaner API. Added parent validation in `Add()` using `findTicketFile()`. Added 9 cobra flags to `cmd/add.go` (`-d/--description`, `-t/--type`, `-p/--priority`, `-a/--assignee`, `--external-ref`, `--parent`, `--design`, `--acceptance`, `--tags`). Default title "Untitled", default type "task", default priority 2, default assignee from `git config user.name`. Description priority: `-d` flag > positional arg > stdin. Updated callers in `cmd/quick_add.go` and `internal/tui/tui.go`. Added 21 new bats tests (51 total). Updated README.md flags table and CHANGELOG.md. All 34 unit tests and 51 bats tests pass.

### Iteration 4: Done sets status=closed and command/test updates
- **Branch**: ralph/done-sets-status-closed
- **PR**: #5 (merged)
- **Summary**: Changed `Done()` from deleting ticket files to setting `status: closed` and writing back to disk. Added closed-ticket filtering in `cmd/list.go` and `internal/tui/tui.go`. Updated `TestDone` and `TestMultipleTickets` in `file_test.go` to verify file persistence and closed status. Rewrote `test/done.bats`: renamed tests to "closes ticket", added `show` assertions for `status: closed`, added new test for file persistence on disk. Updated `README.md` and `CHANGELOG.md`. All 33 unit tests and 30 bats tests pass.
