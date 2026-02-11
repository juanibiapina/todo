# TODO

## Set up release pipeline with GoReleaser and GitHub Actions
---
id: uY3
state: refined
---
Set up automated releases with GoReleaser and GitHub Actions, following the juanibiapina/gob pattern.

**Prerequisites:** Ideally have integration tests (USP) in place first so CI runs them before release.

**1. GitHub Actions workflows:**

`.github/workflows/build-and-test.yaml` — CI on push to main and PRs:
```yaml
on:
  push:
    branches: [main]
  pull_request:
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version-file: go.mod }
      - run: make build
      - run: make test
```

`.github/workflows/release.yaml` — triggered by `v*.*.*` tags:
```yaml
on:
  push:
    tags: ['v*.*.*']
permissions:
  contents: write
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - uses: actions/setup-go@v5
        with: { go-version-file: go.mod }
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
```

**2. GoReleaser config (`.goreleaser.yaml`):**
- Builds: linux/darwin × amd64/arm64, `CGO_ENABLED=0`
- Ldflags: `-X github.com/juanibiapina/todo/internal/version.Version={{.Version}}`
- Archives with README.md + LICENSE.md
- SHA256 checksums
- Changelog: group by conventional commit prefix (feat/fix/others)
- Homebrew tap: `juanibiapina/homebrew-taps` using `HOMEBREW_TAP_TOKEN`

**3. Release docs (`docs/releases.md`):**
- Document tag-based release process
- Semver format: `v1.2.3`, pre-releases: `v1.0.0-beta.1`
- Steps: update CHANGELOG, commit, `git tag v1.x.x`, `git push --tags`

**4. Makefile update:**
- `test` target should run both `unit-test` and `integration-test` (once bats tests exist)

## Space shortcut should paste ticket into calling terminal
---
id: EUZ
state: refined
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
