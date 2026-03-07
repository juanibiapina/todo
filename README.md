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
1 [ ] Buy groceries
2 [ ] Fix the login bug
```

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
- `enter`: toggle checked
- `space`: copy item text to clipboard
- `a`: add new item
- `e`: edit selected item
- `d`: done (check) or delete (if already checked)
- `c`: clean (delete all checked items)
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
