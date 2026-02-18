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
- [ ] Keep 3-char random base62 ID generation (current format). No changes needed to id.go. Verify existing unit tests cover ID uniqueness and format
- [ ] Change file naming from `<id>-<slug>.md` to `<id>.md`. Remove slugify(), update ticketFileName/ticketFilePath, update findTicketFile for exact match, update parseFile/writeFile for YAML-first frontmatter format. Update file_test.go unit tests
- [ ] Update all commands (add, done, show, list, set-description, format) to work with the new file format, naming, and ID generation. Update all existing bats tests to match new output format, ID patterns, and done behavior (status=closed instead of delete)
- [ ] Add creation flags to add command: `-d/--description`, `-t/--type` (bug/feature/task/epic/chore, default task), `-p/--priority` (0-4, default 2), `-a/--assignee` (default git user.name), `--external-ref`, `--parent` (validate exists), `--design`, `--acceptance`, `--tags` (comma-separated). Default title to "Untitled". Add bats tests for each flag and default values
- [ ] Add status management commands: `status <id> <status>` (validate open|in_progress|closed), `start <id>`, `close <id>`, `reopen <id>` shortcuts. Change `done` to set status=closed instead of deleting. Add bats tests for each command, invalid status, and non-existent ticket errors
- [ ] Enhance findTicketFile with partial ID resolution: exact match first, then glob `*<id>*.md`, error on ambiguous matches. Apply to all commands that take an ID. Add bats tests for exact/prefix/suffix/substring matches, ambiguous errors, and exact precedence
- [ ] Add dep/undep commands for dependency management: `dep <id> <dep-id>` (idempotent, validates both exist), `undep <id> <dep-id>`. Store deps as YAML array. Add bats tests for add/remove, idempotency, and validation errors
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
- Existing `file_test.go` tests break because `parseFile()` still reads old format — expected, to be fixed in step 3

## History

### Iteration 1: Ticket struct fields and YAML frontmatter FullString()
- **Branch**: ralph/ticket-yaml-frontmatter
- **PR**: #2 (merged)
- **Summary**: Added 10 new fields to Ticket struct (Status, Type, Priority, Assignee, Created, Parent, ExternalRef, Deps, Links, Tags). Created `frontmatter` helper struct with YAML tags. Rewrote `FullString()` to output `---\nYAML\n---\n# Title\nDescription` format using `gopkg.in/yaml.v3`. Added 12 unit tests in `ticket_test.go`. `String()` method unchanged.
