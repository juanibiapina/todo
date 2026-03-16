# todo

A simple per-directory todo list.

`todo` is a CLI for managing simple todo items stored in a SQLite database. Items are scoped to the directory where you run the command, so each project gets its own list.

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

Items are stored in a SQLite database at `~/.local/share/todo/todo.db` (respects `XDG_DATA_HOME`).

Override the database path with the `TODO_DB` environment variable:

```bash
TODO_DB=/path/to/my.db todo list
```

Items are scoped by the absolute path of the current working directory.

## Development

```bash
make build           # Build binary to dist/todo
make test            # Run all tests (unit + integration)
make unit-test       # Run Go unit tests
make integration-test # Run bats integration tests
```
