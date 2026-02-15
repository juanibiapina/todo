# Space shortcut should paste ticket into calling terminal
---
id: EUZ
---
When running `todo tui` inside a tmux popup, the space shortcut should paste the ticket text directly into the calling tmux pane instead of copying to clipboard.

**Detection:**
- Check `$TMUX` env var to confirm we're in tmux
- Check `$TMUX_PANE` for the current pane ID
- Detect popup context: when in a tmux popup, the popup pane is different from the pane that launched it. Use `tmux display-message -p '#{pane_id}'` to get current pane, and `tmux list-panes -F '#{pane_id}'` on the parent window to find the originating pane. Alternatively, accept a `--tmux-pane <pane-id>` flag that the caller passes in.

**Paste approach:**
- Use `tmux set-buffer` + `tmux paste-buffer -t <target-pane>` to send text to the originating pane
- Or use `tmux send-keys -t <target-pane> -- "<text>"` (simpler but needs escaping)
- `set-buffer` + `paste-buffer` is cleaner for multi-line content

**Changes:**

1. `cmd/tui.go`: add `--tmux-pane` flag (optional, string)
2. `internal/tui/tui.go`:
   - Accept a `tmuxTargetPane` option in `New()` or as a field on `Model`
   - In `copyTicket()`: if `tmuxTargetPane` is set, use `exec.Command("tmux", "set-buffer", "--", text)` then `exec.Command("tmux", "paste-buffer", "-t", targetPane)` instead of clipboard
   - Fall back to clipboard if not in tmux or flag not provided

**Usage:** The user's shell alias/script would call `todo tui --tmux-pane "$TMUX_PANE"` before opening the popup, so the TUI knows where to paste back.
