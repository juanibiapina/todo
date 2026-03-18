# todo

A simple per-directory todo list.

`todo` is a CLI for managing simple todo items stored in a Markdown file. Items are scoped to the directory where you run the command, so each project gets its own list.

## Installation

### Homebrew

```bash
brew install juanibiapina/taps/todo
```

### Go

```bash
go install github.com/juanibiapina/todo@latest
```

### From source

```bash
make build    # builds to dist/todo
make install  # go install
```

## Usage

### Add an item

```bash
todo add Buy groceries
todo add Fix the login bug
```

### Add a section divider

Sections are visual dividers that group items. They can't be checked or cleaned.

```bash
todo add --section
todo add -s
```

### List items

```bash
todo list
```

Or just:

```bash
todo
```

Output:

```
3 [ ] Call dentist  (active)
1 [ ] Buy groceries
2 [ ] Fix the login bug
──────────
```

Active items appear at the top of the list. In the TUI, press `x` to toggle an item as active. Checking an item as done automatically clears its active status.

### Edit an item

```bash
todo edit 1 Buy eggs instead
```

### Check / uncheck items

```bash
todo check 1
todo uncheck 1
```

### Delete checked items

```bash
todo clean
```

### Interactive TUI

```bash
todo tui
```

Keybindings:
- `j`/`k` or arrows: move cursor
- `J`/`K` (Shift): reorder items
- `space`/`enter`: toggle checked
- `x`: toggle active (active items float to top)
- `a`: add new item (after cursor)
- `s`: add section divider (after cursor)
- `e`: edit selected item
- `d`: delete selected item
- `c`: copy item text to clipboard
- `C` (Shift): clean (delete all checked items)
- `/`: filter items (fuzzy search)
- `q`/`esc`: quit

## Storage

Items are stored in a Markdown file. The file is human-readable, version-controllable, and editable with any text editor.

### Markdown format

```markdown
## ~/project-a

- [ ] Fix the login bug
---
- [ ] Deploy to staging

## ~/project-b

- [x] Write tests
- [ ] Update docs @active
```

Each directory gets a `## heading` section. Items use standard Markdown checkboxes (`- [ ]` / `- [x]`). Section dividers are `---`. Active items are marked with `@active`.

### File location

Default location (platform-specific):
- **Linux:** `~/.local/share/todo/todo.md`
- **macOS:** `~/Library/Application Support/todo/todo.md`

The default respects `XDG_DATA_HOME` when set.

### Configuration

Override the file path with the `TODO_FILE` environment variable:

```bash
TODO_FILE=~/todo.md todo list
```

Or create a config file at `<XDG_CONFIG_HOME>/todo/config.toml`:

```toml
file = "~/todo.md"
```

Config values support environment variable expansion (`$HOME`, `${XDG_DATA_HOME}`) and `~` for the home directory.

Resolution order:
1. `TODO_FILE` environment variable
2. Config file `file` key
3. Platform default

## Development

```bash
make build           # Build binary to dist/todo
make test            # Run all tests (unit + integration)
make unit-test       # Run Go unit tests
make integration-test # Run bats integration tests
```
