# TUI Feature Parity Update

## Goal

Update the bubbletea TUI (`internal/tui/`) to reflect all features added across 19 CLI iterations: rich metadata display, computed relationships, view modes (all/ready/blocked/closed), status management actions, edit-in-editor, add-note, add-with-defaults, and updated help. The TUI currently only shows ID+Title in the list and ID+Title+Description in detail, with add (title-only) and done as the only actions.

## Context

**Current TUI state** (`internal/tui/tui.go`, `styles.go`, `scroll.go`, `ansi.go`):
- Split-panel layout: left list panel + right detail viewport (bubbletea `viewport.Model`)
- List shows `ID Title` per line with cursor selection and scroll
- Detail shows Title (bold), ID, and glamour-rendered Description — no other metadata
- Actions: `a` add (title-only modal), `d` done (calls `tickets.Done()`), `space` copy to clipboard
- `loadTickets()` calls `tickets.List()` and filters out `Status == "closed"` — stores only the filtered subset
- Help modal and status bar reflect only the three actions above
- Modals: `modalAdd` (textinput) and `modalHelp` — overlay system via `placeOverlay()`
- Styles in `styles.go`: ANSI 0-15 colors, semantic aliases, lipgloss styles for selection, dialogs, help keys
- `tickMsg` triggers reload every 500ms

**CLI features now available** (from 19 iterations):
- Ticket struct has 15 fields: Title, ID, Description, Status, Type, Priority, Assignee, Created, Parent, ExternalRef, Design, Acceptance, Deps, Links, Tags
- `tickets.ComputeRelations(ticket, allTickets)` returns `TicketRelations` with Blockers, Blocking, Children, Linked, ParentTicket (from `internal/tickets/relations.go`)
- `tickets.SetStatus(dir, id, status)` for status changes (open/in_progress/closed)
- `tickets.AddNote(dir, id, text)` for appending timestamped notes
- `tickets.Add(dir, *Ticket)` accepts full Ticket struct — CLI defaults: type=task, priority=2, assignee=git user.name
- `tea.ExecProcess()` available in bubbletea for suspending TUI to launch external processes like `$EDITOR`
- Ready/blocked/closed filtering logic exists in `cmd/ready.go`, `cmd/blocked.go`, `cmd/closed.go` — uses statusMap pattern over `tickets.List()` results

**Technologies:** Go 1.25, bubbletea/bubbles/lipgloss TUI, glamour markdown rendering, `golang.org/x/term`. Build via `make` (runs `go test` + bats).

**Test baseline:** 93 unit tests, 211 bats integration tests — all passing. TUI is not directly tested by bats but `cmd/tui.go` long description appears in `--help` output.

## Requirements

- All existing tests must continue to pass after each step (`gob run make`)
- TUI detail panel must show all ticket metadata fields and computed relationships (matching CLI `show` output)
- TUI list panel must show status and priority indicators (matching CLI `list`/`ready` output patterns)
- TUI must support switching between view modes: all open, ready, blocked, closed (matching the 4 CLI listing commands)
- Status management (`s` start, `c` close, `r` reopen) must be available from TUI matching CLI `start`/`close`/`reopen`
- Edit-in-editor (`e`) must suspend TUI using `tea.ExecProcess`, matching CLI `edit` behavior
- Add-note (`n`) must use a text input modal, calling `tickets.AddNote()`
- Add modal must apply same defaults as CLI add command (type=task, priority=2, assignee=git user.name)
- Help modal, status bar, and `cmd/tui.go` long description must reflect all new keybindings
- Update README.md "Interactive TUI" section and CHANGELOG.md with TUI enhancements

## Steps

- [x] Enrich the detail panel with all ticket metadata fields. Show Status (display "open" when empty), Type, Priority, Assignee, Created, Parent, ExternalRef, Tags, Deps, Links after the ID line using styled labels. Show Design and Acceptance as labeled sections before Description. Keep existing glamour-rendered Description and markdown cache. Update `updateDetailContent()` in `internal/tui/tui.go`. All existing tests must pass (iteration 1)
- [x] Add computed relationships to the detail panel. Store the full unfiltered ticket list (`allTickets`) on the Model struct alongside the filtered `items`. In `updateDetailContent()`, call `tickets.ComputeRelations(ticket, m.allTickets)` and append Blockers, Blocking, Children, Linked sections (styled headings + `ID [status] Title` lines) after Description. Enhance Parent display with resolved title. Update `loadTickets()` to return both filtered items and all tickets in `ticketsLoadedMsg`. All existing tests must pass (iteration 2)
- [x] Enrich the list panel with status and priority indicators. Update `renderTicketList()` to show `ID [P<n>][status] Title` per line. Color-code priority (P0-P1 red, P2 yellow, P3-P4 muted) and status (in_progress green, open default, closed muted). Adjust `maxTitleLen` for the wider prefix. Update both normal and selected line styles. Add priority/status styles to `styles.go`. All existing tests must pass (iteration 3)
- [x] Add view modes to switch between all-open, ready, blocked, and closed ticket lists. Add `viewMode` enum (`viewAll`, `viewReady`, `viewBlocked`, `viewClosed`) to Model. Add keybindings `1`/`2`/`3`/`4` to switch views. `loadTickets()` stores all tickets; new `applyView()` filters `items` based on `viewMode` using the same logic as `cmd/ready.go` (statusMap, dep checking), `cmd/blocked.go` (unclosed deps), and `cmd/closed.go` (status==closed, sorted by mtime via `os.Stat`). Ready/blocked sort by priority then ID. Show current view in panel title: `"Tickets [All]"`, `"Tickets [Ready]"`, etc. Update status bar with `1-4` hints. All existing tests must pass (iteration 4)
- [ ] Add status management keybindings. In list panel, add `s` (start → `tickets.SetStatus(dir, id, "in_progress")`), `c` (close → `tickets.SetStatus(dir, id, "closed")`), `r` (reopen → `tickets.SetStatus(dir, id, "open")`). Each returns `actionDoneMsg` with message like `"Started: title"`. Update `d` help text from `"done"` to `"close"`. All existing tests must pass
- [ ] Add edit-in-editor action. In list panel, add `e` keybinding that constructs `exec.Command(editor, ticketPath)` using `$EDITOR` (default `vi`) and returns `tea.ExecProcess(cmd, callback)`. Callback returns a message that triggers ticket reload. File path constructed as `tickets.DirPath(dir)/<id>.md`. All existing tests must pass
- [ ] Add add-note modal. Add `modalNote` to `modalMode` enum. In list panel, `n` opens modal with text input (reuse `textInput` pattern from add modal). On enter, call `tickets.AddNote(dir, id, text)` and return `actionDoneMsg`. Modal title: "Add Note", help: "enter: save • esc: cancel". All existing tests must pass
- [ ] Apply CLI defaults in add modal. When creating a ticket via the add modal, construct `tickets.Ticket{Title: title, Type: "task", Priority: 2, Assignee: gitUserName}` where `gitUserName` is resolved from `git config user.name` (same as `cmd/add.go`). Fall back to empty assignee if git command fails. All existing tests must pass
- [ ] Update help modal, status bar, and `cmd/tui.go` long description. Help modal: add `s` start, `c` close, `r` reopen, `e` edit, `n` add note to Ticket List section; change `d` from "mark done (remove)" to "close"; add Views section with `1`/`2`/`3`/`4` keys. Status bar: add key hints for new actions (context-dependent per panel). Update `cmd/tui.go` Long description with complete keybinding reference. All existing tests must pass
- [ ] Update README.md "Interactive TUI" section to document view modes, all keybindings, metadata display, and relationship sections. Add CHANGELOG.md entry under `[Unreleased]` for TUI enhancements. All existing tests must pass

## Learnings

- Design/Acceptance sections render markdown inline without caching (only Description uses the markdown cache) — this keeps complexity low and can be optimized later if needed
- `renderMarkdown()` already exists and works well for rendering Design/Acceptance alongside Description
- `ComputeRelations()` requires the full unfiltered ticket list to resolve relationships correctly — storing `allTickets` separately from the filtered `items` is necessary
- `loadTickets()` can return both filtered and unfiltered lists in the same message struct, keeping the async pattern clean
- Relationship sections reuse existing styles (`ticketIDStyle`, `mutedStyle`, `metaValueStyle`, `sectionHeadingStyle`) — no new styles needed
- Badge helpers (`priorityBadge`, `statusBadge`) need selection-aware background overlay — use `style.Background(selectionBg)` conditional per badge rather than wrapping entire line
- Per-ticket `maxTitleLen` calculation (based on actual status string length) handles variable-width status badges cleanly — `prefixW = 1 + 3 + 1 + 4 + (2 + len(status)) + 1`
- Value receivers work well for read-only helper methods on Model, matching existing patterns like `renderKey`
- Refactoring `loadTickets()` to return all tickets unfiltered and moving filtering to `applyView()` cleanly separates data loading from view logic — view switches don't need disk I/O
- Replicating CLI filtering logic (`cmd/ready.go`, `cmd/blocked.go`, `cmd/closed.go`) in TUI filter methods is straightforward but creates duplication — consider extracting shared filtering functions later
- `os.Stat` on ticket file path for mtime-based sorting in closed view works but couples TUI to filesystem layout (`tickets.DirPath(dir)/ID.md`)

## History

### Iteration 1: Enrich TUI detail panel with all ticket metadata
- **Commit**: 546cf84
- **Summary**: Rewrote `updateDetailContent()` in `internal/tui/tui.go` to display all 15 ticket fields. Added 3 new styles (`metaLabelStyle`, `metaValueStyle`, `sectionHeadingStyle`) in `internal/tui/styles.go`. Status always shown (defaults to "open"), Priority always shown as `P<n>`, other string/slice fields shown conditionally. Design and Acceptance rendered as markdown sections before Description. All 93 unit tests and 211 bats integration tests pass.

### Iteration 2: Add computed relationships to TUI detail panel
- **Commit**: c00c3dc
- **Summary**: Added `allTickets` field to Model struct and updated `ticketsLoadedMsg`/`loadTickets()` to carry both filtered and unfiltered ticket lists. Enhanced Parent display to show resolved `ID (Title)` using `ComputeRelations()`. Added Blockers, Blocking, Children, Linked sections after Description with styled `ID [status] Title` lines. Added `renderRelationSection` and `renderRelationLine` helper methods on `*Model`. All 93 unit tests and 211 bats integration tests pass.

### Iteration 3: Add status and priority badges to TUI list panel
- **Commit**: 1a95414
- **Summary**: Added 6 new style variables (`priorityHighStyle`, `priorityMedStyle`, `priorityLowStyle`, `statusActiveStyle`, `statusDefaultStyle`, `statusClosedStyle`) to `internal/tui/styles.go`. Added `priorityBadge()` and `statusBadge()` helper methods on Model. Rewrote `renderTicketList()` loop to display `ID [P<n>][status] Title` with color-coded priority (P0-P1 red, P2 yellow, P3+ muted) and status (in_progress green, open default, closed muted) badges. Dynamic per-ticket `maxTitleLen` based on status string length. All 93 unit tests and 211 bats integration tests pass.

### Iteration 4: Add view modes to TUI for filtering tickets
- **Commit**: 7205641
- **Summary**: Added `viewMode` type with 4 constants (`viewAll`, `viewReady`, `viewBlocked`, `viewClosed`) and `view` field to Model. Refactored `loadTickets()` to return all tickets unfiltered; added `applyView()` method that dispatches to `filterReady()`, `filterBlocked()`, `filterClosed()` matching CLI logic. Ready/blocked sort by priority then ID; closed sorts by file mtime descending. Added `1`/`2`/`3`/`4` keybindings in `updateListPanel()` to switch views with scroll reset. Dynamic list panel title shows current view. Status bar shows `1-4 views` hint. All 93 unit tests and 211 bats integration tests pass.
