# todo

Local ticket tracking in markdown.

`todo` is a CLI for managing tickets stored as a `TODO.md` file in your project directory. Each ticket has a title, a 3-character ID, and an optional description.

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
# Simple ticket
todo add 'Fix login timeout'

# With inline description
todo add 'Fix login timeout' 'Users experience timeouts after 30s'

# With rich description via stdin
todo add 'Fix login timeout' <<'EOF'
Users experience timeouts after 30s on the login page.

The `handleLogin` function in `auth.go` needs a longer timeout:

```go
ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
```
EOF
````

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

### Reorder tickets

```bash
todo move-up aBc
todo move-down aBc
```

### Complete a ticket

```bash
todo done aBc
```

### Interactive TUI

```bash
todo tui
```

Launches a full-screen terminal interface with a split-panel layout (ticket list + detail view). Navigate, add, delete, and reorder tickets interactively.

### Quick add (for tmux popups)

```bash
todo quick-add
```

Opens an interactive prompt to quickly add a ticket and exit.

## File Format

Tickets are stored in `TODO.md` in the current directory:

```markdown
# TODO

## Fix login timeout
---
id: aBc
---

## Refactor auth module
---
id: xYz
---
Move auth logic to middleware layer.
Multiple lines of description are supported.
```

Each `##` heading is a ticket title. A YAML front matter block (`---` delimited) follows with `id`. Everything after the closing `---` until the next `##` is the description.

## License

MIT
