# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-02-11

### Added

- `todo add` - Create tickets with optional description (via argument or stdin)
- `todo list` - List all tickets with ID, state, and title
- `todo show` - Show full ticket details by ID or title
- `todo done` - Remove a ticket (mark complete)
- `todo set-state` - Change ticket state (new, refined, planned)
- `todo set-description` - Set/replace description (via argument or stdin)
- Stdin support for descriptions: heredocs and pipes work with backticks, code blocks, and special characters
- Tickets stored in `.tickets.md` in the current directory
- 3-character base62 ticket IDs
