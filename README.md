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
# aBc Fix login timeout
# xYz Refactor auth module
```

IDs are color-coded (magenta) when output is a terminal.

### Show a ticket

```bash
todo show aBc
```

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

### Interactive TUI

```bash
todo tui
```

Launches a full-screen terminal interface with a split-panel layout (ticket list + detail view). Navigate, add, and close tickets interactively.

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
