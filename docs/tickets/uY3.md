---
id: uY3
---
# Set up release pipeline with GoReleaser and GitHub Actions
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
