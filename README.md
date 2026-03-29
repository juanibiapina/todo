# todo

A simple per-directory todo list.

`todo` is a CLI for managing simple todo items stored in a Markdown file. Items are scoped to the directory where you run the command, so each project gets its own list.

![demo](assets/demo.gif)

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
2   [ ] Pick up milk
──────────
```

Active items appear at the top of the list. Indented items show extra spacing before the checkbox. In the TUI, press `x` to toggle an item as active. Checking an item as done automatically clears its active status.

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
- `tab`: indent selected item
- `shift+tab`: unindent selected item
- `a`: add new item (after cursor, no indent)
- `o`: add new item (after cursor, matches indent level)
- `O` (Shift): add new item (before cursor, inherits indent)
- `tab`/`shift+tab` in add/edit mode: adjust indent level
- `esc` in add/edit mode: save (same as `enter`)
- `ctrl+c` in add/edit mode: cancel
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
  - [ ] Reproduce locally
    - [ ] Check server logs
---
- [ ] Deploy to staging

## ~/project-b

- [x] Write tests
- [ ] Update docs @active
```

Each directory gets a `## heading` section. Items use standard Markdown checkboxes (`- [ ]` / `- [x]`). Indented items use nested list syntax (2 spaces per level, up to 5 levels). Section dividers are `---`. Active items are marked with `@active`.

### File location

Default: `~/.local/share/todo/todo.md`

Override with the `TODO_FILE` environment variable:

```bash
export TODO_FILE=~/todo.md
```

The value supports environment variable expansion (`$HOME`, `$MY_VAR`) and `~` for the home directory.

## Development

```bash
make build           # Build binary to dist/todo
make test            # Run all tests (unit + integration)
make unit-test       # Run Go unit tests
make integration-test # Run bats integration tests
```
