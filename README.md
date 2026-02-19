# todo

Local ticket tracking in markdown.

`todo` is a CLI for managing tickets stored in a `docs/tickets/` directory in your project. Each ticket is a separate markdown file with a title, a 3-character ID, and an optional description.

Descriptions can be passed via stdin (heredocs, pipes) so multi-line content with backticks, code blocks, and special characters works without shell escaping issues.

## Installation

```bash
go install github.com/juanibiapina/todo@latest
```

Or build from source:

```bash
make build    # builds to dist/todo
make install  # go install
```

## Usage

### Add a ticket

````bash
# Simple ticket (defaults: type=task, priority=2, assignee=git user.name)
todo add 'Fix login timeout'

# With inline description
todo add 'Fix login timeout' 'Users experience timeouts after 30s'

# With description flag
todo add 'Fix login timeout' -d 'Users experience timeouts after 30s'

# With rich description via stdin
todo add 'Fix login timeout' <<'EOF'
Users experience timeouts after 30s on the login page.

The `handleLogin` function in `auth.go` needs a longer timeout:

```go
ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
```
EOF

# With metadata flags
todo add 'Fix login timeout' -t bug -p 3 -a 'Alice' --tags 'auth,urgent'

# With parent ticket and external reference
todo add 'Subtask of login fix' --parent aBc --external-ref JIRA-456

# With design and acceptance criteria
todo add 'Refactor auth' --design 'Extract token validation' --acceptance 'All tests pass'

# Default title ("Untitled") when no args
todo add -t epic -p 1
````

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--description` | `-d` | | Ticket description (overrides positional arg and stdin) |
| `--type` | `-t` | `task` | Type: `bug`, `feature`, `task`, `epic`, `chore` |
| `--priority` | `-p` | `2` | Priority: `0`–`4` |
| `--assignee` | `-a` | `git user.name` | Assignee |
| `--external-ref` | | | External reference (e.g. JIRA-123) |
| `--parent` | | | Parent ticket ID (must exist) |
| `--design` | | | Design notes |
| `--acceptance` | | | Acceptance criteria |
| `--tags` | | | Comma-separated tags |

### List tickets

```bash
todo list
# aBc - Fix login timeout
# xYz [in_progress] - Refactor auth module <- [qRs, mNp]
```

Output format: `id [status] - Title <- [dep1, dep2]`. The `[status]` portion is omitted when empty, and `<- [deps]` is omitted when there are no dependencies. IDs are color-coded (magenta) when output is a terminal.

By default, closed tickets are hidden. Use `--status` to filter by status:

```bash
# Show only in-progress tickets
todo list --status in_progress

# Show closed tickets
todo list --status closed

# Show open tickets (includes newly created tickets with no status set)
todo list --status open
```

Filter by assignee or tag:

```bash
# Filter by assignee
todo list -a Alice

# Filter by tag
todo list -T urgent

# Combine filters (AND logic)
todo list --status open -a Alice -T auth
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--status` | | Filter by status: `open`, `in_progress`, `closed` |
| `--assignee` | `-a` | Filter by assignee |
| `--tag` | `-T` | Filter by tag |

An empty result produces no output.

### Show a ticket

```bash
todo show aBc
```

The output includes YAML frontmatter, title, description, and computed relationship sections:

- **Blockers** — unclosed tickets this ticket depends on
- **Blocking** — tickets that depend on this ticket (when this ticket is unclosed)
- **Children** — tickets whose parent is this ticket
- **Linked** — tickets linked to this ticket

Sections are only shown when non-empty. Each entry is formatted as `- id [status] Title`.

When a ticket has a parent, the frontmatter `parent:` line is enhanced with the parent's title (e.g. `parent: aBc (Fix login timeout)`).

#### Pager support

Set the `TODO_PAGER` environment variable to pipe `show` output through a pager:

```bash
export TODO_PAGER="less -R"
todo show aBc
```

The pager is only used when stdout is a terminal (piped output is never paged).

### Partial ID matching

All commands that accept a ticket ID support partial matching. You can use any unique substring of the ID:

```bash
# These all work if "aBc" is the only ticket containing "aB"
todo show aB
todo done aB
todo start aB
todo status aB in_progress
todo set-description aB 'Updated description'
```

Exact matches always take precedence. If a partial ID matches multiple tickets, an error lists the ambiguous IDs.

### Set description

```bash
# Simple
todo set-description aBc 'New description'

# Rich content via stdin
todo set-description aBc <<'EOF'
### Problem
The `validateToken` function doesn't handle expired tokens.

### Plan
1. Add expiry check in `validateToken`
2. Return proper error type
3. Update callers
EOF
```

### Edit a ticket

```bash
todo edit aBc
```

Opens the ticket file in your `$EDITOR` (defaults to `vi` if not set). Supports partial ID matching.

When stdout is not a terminal (e.g. in a script or pipe), the file path is printed instead of launching an editor:

```bash
# Get the file path for scripting
path=$(todo edit aBc)
echo "$path"  # docs/tickets/aBc.md
```

### Complete a ticket

```bash
todo done aBc
```

This sets the ticket's status to `closed`. The ticket file is preserved on disk but hidden from `list` and the TUI. Use `show` to view closed tickets.

### Manage ticket status

Valid statuses: `open`, `in_progress`, `closed`.

```bash
# Set status directly
todo status aBc in_progress

# Shortcut commands
todo start aBc      # set status to in_progress
todo close aBc      # set status to closed
todo reopen aBc     # set status to open
```

Closed tickets are hidden from `list` and the TUI. Use `reopen` to make them visible again.

### Manage dependencies

```bash
# Add a dependency (ticket aBc depends on xYz)
todo dep aBc xYz

# Remove a dependency
todo undep aBc xYz
```

Both commands validate that the referenced tickets exist. Operations are idempotent — adding an existing dependency or removing a non-existent one succeeds silently. Partial ID matching is supported for both arguments.

#### Dependency tree

```bash
# Show dependency tree
todo dep tree aBc
# aBc Fix login timeout
# ├── xYz [in_progress] Refactor auth module
# │   └── qRs Validate tokens
# └── mNp Write tests

# Show full tree (no deduplication)
todo dep tree aBc --full
```

The tree uses box-drawing characters (`├── `, `└── `, `│   `) and shows `[status]` when set. Children are sorted by subtree depth (deepest first), then by ID. Cycles are marked with `(cycle)` and duplicate nodes with `(dup)`. Use `--full` to disable deduplication.

#### Cycle detection

```bash
# Detect dependency cycles among open tickets
todo dep cycle
```

Performs DFS-based cycle detection on all non-closed tickets. Each cycle is displayed as a normalized path (rotated so the smallest ID is first) with member details:

```
Cycle: aBc -> xYz -> aBc
  aBc [open] Fix login timeout
  xYz [in_progress] Refactor auth module
```

Closed tickets are excluded from the analysis. If no cycles are found, the command produces no output.

### Ready tickets

```bash
todo ready
# aBc [P1][open] - Critical bug fix
# xYz [P2][in_progress] - Refactor auth module
```

Shows tickets that are ready to work on: open or in-progress tickets where all dependencies are closed (or have no dependencies). Missing deps are treated as non-blocking.

Output format: `id [P<priority>][status] - Title`. Priority is always shown. Empty status is displayed as `[open]`. Tickets are sorted by priority ascending (lower number = higher priority), then by ID.

Filter by assignee or tag:

```bash
# Filter by assignee
todo ready -a Alice

# Filter by tag
todo ready -T urgent
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--assignee` | `-a` | Filter by assignee |
| `--tag` | `-T` | Filter by tag |

### Blocked tickets

```bash
todo blocked
# aBc [P1][open] - Critical bug fix <- [xYz]
# mNp [P2][in_progress] - Refactor auth module <- [qRs, jKl]
```

Shows tickets that are blocked: open or in-progress tickets with at least one unclosed dependency. Only the unclosed blockers are shown in the output. Missing deps (not found in the ticket list) are treated as non-blocking.

Output format: `id [P<priority>][status] - Title <- [blocker1, blocker2]`. Priority is always shown. Empty status is displayed as `[open]`. Tickets are sorted by priority ascending (lower number = higher priority), then by ID.

Filter by assignee or tag:

```bash
# Filter by assignee
todo blocked -a Alice

# Filter by tag
todo blocked -T urgent
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--assignee` | `-a` | Filter by assignee |
| `--tag` | `-T` | Filter by tag |

### Closed tickets

```bash
todo closed
# xYz - Completed login fix
# aBc - Old feature request
```

Shows recently closed tickets sorted by file modification time (most recent first). Default limit is 20.

Output format: `id - Title`. Status is omitted since all displayed tickets are closed.

```bash
# Limit the number of results
todo closed --limit 10
todo closed -n 5

# Filter by assignee
todo closed -a Alice

# Filter by tag
todo closed -T backend
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--limit` | `-n` | `20` | Maximum number of tickets to show |
| `--assignee` | `-a` | | Filter by assignee |
| `--tag` | `-T` | | Filter by tag |

### Query tickets (JSONL)

```bash
todo query
# {"id":"aBc","title":"Fix login timeout","priority":2,"deps":[],"links":[],"tags":[]}
# {"id":"xYz","title":"Refactor auth","status":"in_progress","type":"bug","priority":1,"assignee":"Alice","deps":["aBc"],"links":[],"tags":["auth","urgent"]}
```

Outputs all tickets as JSON Lines (one JSON object per line) with all frontmatter fields. Useful for scripting, external tooling, and data analysis with `jq`:

```bash
# Count open tickets
todo query --status open | wc -l

# Get all high-priority tickets
todo query | jq -r 'select(.priority <= 1) | .title'

# Export to a file
todo query > tickets.jsonl
```

Fields `id`, `title`, and `priority` are always present. String fields are omitted when empty. Array fields (`deps`, `links`, `tags`) are always present as arrays (never `null`).

Filter by status, type, assignee, or tag:

```bash
# Filter by status
todo query --status open

# Filter by type
todo query --type bug

# Filter by assignee
todo query -a Alice

# Filter by tag
todo query -T urgent

# Combine filters (AND logic)
todo query --status open --type bug -a Alice
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--status` | | Filter by status: `open`, `in_progress`, `closed` |
| `--type` | | Filter by type: `bug`, `feature`, `task`, `epic`, `chore` |
| `--assignee` | `-a` | Filter by assignee |
| `--tag` | `-T` | Filter by tag |

### Add notes

```bash
# Add a note with positional argument
todo add-note aBc 'Discussed with team, agreed on approach'

# Add a note via stdin (for multi-line content)
todo add-note aBc <<'EOF'
Spoke with PM about requirements:
- Need to support OAuth2
- Deadline is next Friday
EOF
```

Notes are appended to the ticket's description under a `## Notes` section. Each note is prefixed with a UTC timestamp in bold:

```markdown
## Notes

**2026-02-18 10:30 UTC**

Discussed with team, agreed on approach

**2026-02-18 14:15 UTC**

Follow-up: implementation started
```

The `## Notes` header is created automatically on the first note and reused for subsequent notes. Partial ID matching is supported.

### Manage links

```bash
# Link two tickets bidirectionally (both files updated)
todo link aBc xYz

# Link three or more tickets (all pairs linked)
todo link aBc xYz qRs

# Remove a link between two tickets (both sides removed)
todo unlink aBc xYz
```

Both commands validate that the referenced tickets exist. Operations are idempotent — adding an existing link or removing a non-existent one succeeds silently. Partial ID matching is supported.

### Interactive TUI

```bash
todo tui
```

Launches a full-screen terminal interface with a split-panel layout for managing tickets interactively.

**Panels:**

- **List panel** (left) — Shows tickets as `ID [P<n>][status] Title` with color-coded badges. Priority: P0–P1 red, P2 yellow, P3+ muted. Status: `in_progress` green, `open` default, `closed` muted.
- **Detail panel** (right) — Shows full ticket metadata (Status, Type, Priority, Assignee, Created, Parent, Ref, Tags, Deps, Links), markdown-rendered Design/Acceptance/Description sections, and computed relationships (Blockers, Blocking, Children, Linked).

**View modes:**

| Key | View | Description |
|-----|------|-------------|
| `1` | All | Open and in-progress tickets (default) |
| `2` | Ready | Tickets with all deps closed or no deps |
| `3` | Blocked | Tickets with at least one unclosed dep |
| `4` | Closed | Closed tickets sorted by last modified |

**Keybindings:**

| Context | Key | Action |
|---------|-----|--------|
| List | `↑`/`k`, `↓`/`j` | Move cursor |
| List | `g`/`G` | First / last ticket |
| List | `a` | Add ticket (defaults: type=task, priority=2, assignee=git user.name) |
| List | `s` | Start ticket (set status to `in_progress`) |
| List | `c`/`d` | Close ticket (set status to `closed`) |
| List | `r` | Reopen ticket (set status to `open`) |
| List | `e` | Edit ticket in `$EDITOR` |
| List | `n` | Add note to ticket |
| List | `space` | Copy ticket ID to clipboard |
| Detail | `↑`/`k`, `↓`/`j` | Scroll content |
| Detail | `g`/`G` | Top / bottom |
| Detail | `ctrl+u`/`ctrl+d` | Half page up / down |
| General | `tab` | Switch panels |
| General | `?` | Show help |
| General | `esc`/`q` | Quit |

### Quick add (for tmux popups)

```bash
todo quick-add
```

Opens an interactive prompt to quickly add a ticket and exit.

## File Format

Tickets are stored in a `docs/tickets/` directory, one file per ticket. Each file is named `<id>.md`:

```
docs/tickets/
├── aBc.md
└── xYz.md
```

Each file contains YAML frontmatter followed by the title and optional description:

```markdown
---
id: aBc
type: task
priority: 2
assignee: Alice
---
# Fix login timeout
Description goes here.
Multiple lines are supported.
```

The YAML frontmatter block (`---` delimited) contains the ticket metadata. The `id` field is always present. Other fields (`status`, `type`, `priority`, `assignee`, `external_ref`, `parent`, `design`, `acceptance`, `tags`, `deps`, `links`, `created`) are included only when set (empty values are omitted). The `# Title` heading follows the frontmatter. Everything after the title line is the description.

## License

MIT
