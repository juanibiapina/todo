//! Styles for the TUI. Every colour is one of the terminal's 16 palette
//! slots, so the UI takes its look from the terminal theme (for example
//! tokyonight moon) and stays readable on light and dark themes.

use ratatui::style::{Color, Modifier, Style};

/// Selection marker, the filter prompt, and the help popup title.
pub const ACCENT: Style = Style::new().fg(Color::Magenta).add_modifier(Modifier::BOLD);

/// The directory in the header and section titles.
pub const TITLE: Style = Style::new().fg(Color::Blue).add_modifier(Modifier::BOLD);

/// Key names in hints and on the help screen.
pub const KEY: Style = Style::new().fg(Color::Blue).add_modifier(Modifier::BOLD);

/// Group headings on the help screen.
pub const HEADING: Style = Style::new().fg(Color::Cyan).add_modifier(Modifier::BOLD);

/// Secondary text: counts, hint labels, open checkboxes, placeholders.
pub const MUTED: Style = Style::new().fg(Color::Gray);

/// Decoration that should recede: section rules.
pub const FAINT: Style = Style::new().fg(Color::DarkGray);

/// The help popup border.
pub const BORDER: Style = Style::new().fg(Color::Blue);

/// Text of checked items.
pub const DONE: Style = Style::new()
    .fg(Color::Gray)
    .add_modifier(Modifier::CROSSED_OUT);

/// The checkbox of checked items.
pub const CHECKED: Style = Style::new().fg(Color::Green);

/// Items marked active.
pub const ACTIVE: Style = Style::new().fg(Color::Yellow).add_modifier(Modifier::BOLD);

/// Characters matched by the filter.
pub const MATCH: Style = Style::new().fg(Color::Cyan).add_modifier(Modifier::BOLD);

/// Confirmation messages in the footer.
pub const SUCCESS: Style = Style::new().fg(Color::Green);

/// Failed operations.
pub const ERROR: Style = Style::new().fg(Color::Red).add_modifier(Modifier::BOLD);
