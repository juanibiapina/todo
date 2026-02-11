# todo

Local ticket tracking in markdown.

`todo` is a CLI for managing tickets stored as a `.tickets.md` file in your project directory. Each ticket has a title, a 3-character ID, a state, and an optional description.

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

```bash
# Simple ticket (state: new)
todo add 'Fix login timeout'

# With inline description (state: refined)
todo add 'Fix login timeout' 'Users experience timeouts after 30s'

# With rich description via stdin
todo add 'Fix login timeout' <<'EOF'
Users experience timeouts after 30s on the login page.

The `handleLogin` function in `auth.go` needs a longer timeout:

```go
ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
```
EOF
```

### List tickets

```bash
todo list
# aBc [new     ] Fix login timeout
# xYz [refined ] Refactor auth module
```

### Show a ticket

```bash
todo show aBc
todo show 'Fix login timeout'
```

### Set state

```bash
todo set-state aBc refined
todo set-state aBc planned
```

Valid states: `new`, `refined`, `planned`

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

## File Format

Tickets are stored in `.tickets.md` in the current directory:

```markdown
# Tickets

## Fix login timeout
---
id: aBc
state: new
---

## Refactor auth module
---
id: xYz
state: refined
---
Move auth logic to middleware layer.
Multiple lines of description are supported.

## Planned ticket
---
id: Qr5
state: planned
---
### Problem
Description of the problem.

### Plan
1. Step one
2. Step two
```

Each `##` heading is a ticket title. A YAML front matter block (`---` delimited) follows with `id` and `state`. Everything after the closing `---` until the next `##` is the description.

## States

| State | Meaning |
|-------|---------|
| `new` | Just created, no description |
| `refined` | Has a description |
| `planned` | Analyzed, has execution plan. Ready to implement |

## License

MIT
