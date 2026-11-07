//! Rendering: turns the App state into lines for the content area.

use ratatui::Frame;
use ratatui::layout::{Position, Rect};
use ratatui::text::{Line, Span};
use ratatui::widgets::{Block, Clear, Padding, Paragraph};
use unicode_width::UnicodeWidthStr;

use super::theme::{
    ACCENT, ACTIVE, BORDER, CHECKED, DONE, ERROR, FAINT, HEADING, KEY, MATCH, MUTED, SUCCESS, TITLE,
};
use super::{App, FOOTER_LINES, FRAME_PADDING_X, FRAME_PADDING_Y, HEADER_LINES, Mode, wrap};
use crate::render::section_parts;
use crate::store::Item;

/// A key and what it does, shown in the footer.
type Hint = (&'static str, &'static str);

/// Normal-mode hints in priority order; the footer drops them from the end
/// when the terminal is narrow. `NORMAL_PINNED` is always shown.
const NORMAL_HINTS: &[Hint] = &[
    ("a", "add"),
    ("space", "done"),
    ("e", "edit"),
    ("x", "active"),
    ("/", "filter"),
];
const NORMAL_PINNED: &[Hint] = &[("?", "help"), ("q", "quit")];
const ADD_HINTS: &[Hint] = &[
    ("enter", "save"),
    ("tab", "indent"),
    ("shift+tab", "outdent"),
    ("ctrl+c", "cancel"),
];
const EDIT_HINTS: &[Hint] = &[("enter", "save"), ("ctrl+c", "cancel")];
const FILTER_HINTS: &[Hint] = &[("↑↓", "move"), ("enter", "jump"), ("esc", "cancel")];
const HELP_HINTS: &[Hint] = &[("? esc", "close")];
const HELP_SCROLL_HINTS: &[Hint] = &[("j k", "scroll"), ("? esc", "close")];

/// Full keymap for the help screen, grouped by purpose.
const KEYMAP: &[(&str, &[Hint])] = &[
    (
        "Move",
        &[
            ("j k ↓ ↑", "move the cursor"),
            ("J K", "move the item down or up"),
            ("/", "filter items and jump to one"),
        ],
    ),
    (
        "Add and edit",
        &[
            ("a", "add an item below"),
            ("o", "add an item below at the same depth"),
            ("O", "add an item above"),
            ("s", "add a section divider below"),
            ("e", "edit the item or section title"),
            ("d", "delete the item"),
            ("c", "copy the item text"),
        ],
    ),
    (
        "Organise",
        &[
            ("space enter", "mark done or not done"),
            ("x", "mark active; active items stay on top"),
            ("tab shift+tab", "indent or outdent"),
            ("C", "remove all done items"),
        ],
    ),
    ("Other", &[("?", "show this help"), ("q esc", "quit")]),
];

/// Width of the key column on the help screen.
const KEY_COLUMN: usize = 15;
/// Width of one help column: keys plus the longest description.
const HELP_COLUMN: usize = 2 + KEY_COLUMN + 38;
/// Border and padding around the help popup text.
const HELP_CHROME_X: usize = 4;
const HELP_CHROME_Y: usize = 2;

/// Rendered content plus the terminal cursor position, both relative to the
/// content area.
struct View {
    lines: Vec<Line<'static>>,
    cursor: Option<(usize, usize)>,
}

impl App {
    pub(super) fn draw(&mut self, frame: &mut Frame) {
        let area = frame.area();
        let content = Rect {
            x: area.x + FRAME_PADDING_X.min(area.width),
            y: area.y + FRAME_PADDING_Y.min(area.height),
            width: area.width.saturating_sub(2 * FRAME_PADDING_X),
            height: area.height.saturating_sub(2 * FRAME_PADDING_Y),
        };

        let view = match self.mode {
            Mode::Filter => self.view_filter(),
            _ => self.view_list(),
        };
        frame.render_widget(Paragraph::new(view.lines), content);
        if let Some((row, col)) = view.cursor
            && row < content.height as usize
            && col < content.width as usize
        {
            frame.set_cursor_position(Position {
                x: content.x + col as u16,
                y: content.y + row as u16,
            });
        }
        if self.mode == Mode::Help {
            self.draw_help(frame, content);
        }
    }

    /// Draws the keymap in a bordered popup centred over the list, between
    /// the header and the footer.
    fn draw_help(&self, frame: &mut Frame, content: Rect) {
        let content = Rect {
            y: content.y + HEADER_LINES as u16,
            height: content
                .height
                .saturating_sub((HEADER_LINES + FOOTER_LINES) as u16),
            ..content
        };
        let lines = self.help_lines();
        let rows = self.help_rows();
        let scroll = self.help_scroll.min(lines.len().saturating_sub(rows));
        let more_below = scroll + rows < lines.len();

        let text_width = lines.iter().map(Line::width).max().unwrap_or(0);
        let width = (text_width + HELP_CHROME_X).min(content.width as usize) as u16;
        let height = (rows + HELP_CHROME_Y) as u16;
        let area = Rect {
            x: content.x + (content.width.saturating_sub(width)) / 2,
            y: content.y + (content.height.saturating_sub(height)) / 2,
            width,
            height: height.min(content.height),
        };

        let mut block = Block::bordered()
            .border_style(BORDER)
            .title(Span::styled(" Keys ", ACCENT))
            .padding(Padding::horizontal(1));
        if more_below {
            block = block.title_bottom(Line::styled(" j more ", MUTED).right_aligned());
        }

        frame.render_widget(Clear, area);
        frame.render_widget(
            Paragraph::new(lines.into_iter().skip(scroll).collect::<Vec<_>>()).block(block),
            area,
        );
    }

    /// The directory, then a blank line.
    fn header(&self) -> Vec<Line<'static>> {
        let path = truncate_left(&self.display_dir, self.width);
        vec![Line::styled(path, TITLE), Line::raw("")]
    }

    /// A blank line and either the status message or the key hints.
    fn footer(&self, hints: &[Hint]) -> Vec<Line<'static>> {
        let pinned = if self.mode == Mode::Normal {
            NORMAL_PINNED
        } else {
            &[]
        };
        let last = match &self.status {
            Some(status) => {
                let style = if status.is_error { ERROR } else { SUCCESS };
                Line::styled(status.text.clone(), style)
            }
            None => hint_line(hints, pinned, self.width),
        };
        vec![Line::raw(""), last]
    }

    fn view_list(&mut self) -> View {
        let mut lines: Vec<Line<'static>> = Vec::new();
        let mut input_at: Option<(usize, usize)> = None;

        if self.items.is_empty() && self.mode == Mode::Add {
            input_at = Some(self.push_add_input(&mut lines));
        } else if self.items.is_empty() {
            lines.extend(empty_state());
        }

        let gap = self.active_divider_index();
        for i in 0..self.items.len() {
            // A blank line separates active items from the rest.
            if gap == Some(i) && !self.items[i].is_section {
                lines.push(Line::raw(""));
            }
            let add_here = self.mode == Mode::Add && i == self.cursor;
            if add_here && self.add_before {
                input_at = Some(self.push_add_input(&mut lines));
            }
            if self.mode == Mode::Edit && i == self.cursor {
                input_at = Some(self.push_edit_input(&mut lines, i));
            } else {
                lines.extend(self.item_lines(i));
            }
            if add_here && !self.add_before {
                input_at = Some(self.push_add_input(&mut lines));
            }
        }

        self.frame_list(lines, input_at)
    }

    /// Clips the item lines to the viewport and adds the header and footer.
    fn frame_list(&self, mut lines: Vec<Line<'static>>, input_at: Option<(usize, usize)>) -> View {
        let available = if self.height > 0 {
            self.height
                .saturating_sub(HEADER_LINES + FOOTER_LINES)
                .max(1)
        } else {
            lines.len()
        };
        let mut start = self.scroll_offset;
        if start > lines.len() {
            start = lines.len().saturating_sub(available);
        }
        let end = (start + available).min(lines.len());

        let cursor = input_at
            .filter(|(row, _)| (start..end).contains(row))
            .map(|(row, col)| (HEADER_LINES + row - start, col));

        let hints = match self.mode {
            Mode::Add => ADD_HINTS,
            Mode::Edit => EDIT_HINTS,
            Mode::Help if self.help_lines().len() > self.help_rows() => HELP_SCROLL_HINTS,
            Mode::Help => HELP_HINTS,
            _ => NORMAL_HINTS,
        };

        let mut out = self.header();
        out.extend(lines.drain(start..end));
        // Keep the footer at the bottom of the screen.
        let used = out.len() + FOOTER_LINES;
        out.extend((used..self.height).map(|_| Line::raw("")));
        out.extend(self.footer(hints));
        View { lines: out, cursor }
    }

    /// Builds an input line: marker + indent + `before` + input. Returns the
    /// line and the cursor column.
    fn input_line(&mut self, indent: &str, before: Span<'static>) -> (Line<'static>, usize) {
        let before_width = before.width();
        let mut spans = vec![
            Span::styled("› ", ACCENT),
            Span::raw(indent.to_string()),
            before,
        ];
        let (input_spans, col) = self.input.render();
        spans.extend(input_spans);
        (Line::from(spans), 2 + indent.len() + before_width + col)
    }

    /// Pushes the add-mode input line and returns its row and cursor column.
    fn push_add_input(&mut self, lines: &mut Vec<Line<'static>>) -> (usize, usize) {
        let indent = "  ".repeat(self.add_indent);
        let (line, col) = self.input_line(&indent, Span::raw("[ ] "));
        lines.push(line);
        (lines.len() - 1, col)
    }

    /// Pushes the edit input that replaces item `i` and returns its row and
    /// cursor column.
    fn push_edit_input(&mut self, lines: &mut Vec<Line<'static>>, i: usize) -> (usize, usize) {
        let item = &self.items[i];
        let (line, col) = if item.is_section {
            self.input_line("", Span::styled("── ", FAINT))
        } else {
            let indent = "  ".repeat(item.indent_level);
            let check = Span::raw(format!("{} ", check_box(item)));
            self.input_line(&indent, check)
        };
        lines.push(line);
        (lines.len() - 1, col)
    }

    fn marker(&self, i: usize) -> Span<'static> {
        if i == self.cursor && self.mode != Mode::Add {
            Span::styled("› ", ACCENT)
        } else {
            Span::raw("  ")
        }
    }

    /// Lines for item `i`: one for a section, or the wrapped item text.
    fn item_lines(&self, i: usize) -> Vec<Line<'static>> {
        let item = &self.items[i];
        let marker = self.marker(i);

        if item.is_section {
            let width = if self.width > 0 {
                self.width.saturating_sub(2)
            } else {
                10
            };
            let (lead, mid, tail) = section_parts(&item.text, width);
            return vec![Line::from(vec![
                marker,
                Span::styled(lead, FAINT),
                Span::styled(mid, TITLE),
                Span::styled(tail, FAINT),
            ])];
        }

        let check = Span::styled(format!("{} ", check_box(item)), check_style(item));
        let style = item_style(item);
        let indent = "  ".repeat(item.indent_level);
        let prefix_width = 2 + indent.len() + 4;

        if self.width == 0 || prefix_width >= self.width {
            return vec![Line::from(vec![
                marker,
                check,
                Span::styled(item.text.clone(), style),
            ])];
        }

        let continuation = " ".repeat(prefix_width);
        wrap(&item.text, self.width - prefix_width)
            .into_iter()
            .enumerate()
            .map(|(j, text)| {
                let mut spans = if j == 0 {
                    vec![marker.clone(), Span::raw(indent.clone()), check.clone()]
                } else {
                    vec![Span::raw(continuation.clone())]
                };
                spans.push(Span::styled(text, style));
                Line::from(spans)
            })
            .collect()
    }

    fn view_filter(&mut self) -> View {
        let mut out = self.header();

        let (mut input_spans, col) = self.filter_input.render();
        input_spans[0].style = ACCENT;
        let count = format!(
            "{} of {}",
            self.filter_matches.len(),
            self.filter_item_indexes.len()
        );
        let used: usize = input_spans.iter().map(Span::width).sum();
        let gap = self.width.saturating_sub(used + count.width()).max(2);
        input_spans.push(Span::raw(" ".repeat(gap)));
        input_spans.push(Span::styled(count, MUTED));
        let cursor = Some((out.len(), col));
        out.push(Line::from(input_spans));

        if self.filter_matches.is_empty() {
            out.push(Line::styled("  No items match.", MUTED));
        }
        for i in self.filter_window() {
            out.push(self.filter_line(i));
        }

        let used = out.len() + FOOTER_LINES;
        out.extend((used..self.height).map(|_| Line::raw("")));
        out.extend(self.footer(FILTER_HINTS));
        View { lines: out, cursor }
    }

    /// Range of filter matches that fits on screen and contains the cursor.
    fn filter_window(&self) -> std::ops::Range<usize> {
        let total = self.filter_matches.len();
        // Header (2) + filter input (1).
        let available = if self.height > 0 {
            self.height
                .saturating_sub(HEADER_LINES + 1 + FOOTER_LINES)
                .max(1)
        } else {
            total
        };

        let mut offset = self.scroll_offset.min(self.filter_cursor);
        if self.filter_cursor >= offset + available {
            offset = self.filter_cursor + 1 - available;
        }
        offset = offset.min(total.saturating_sub(available));
        offset..(offset + available).min(total)
    }

    fn filter_line(&self, i: usize) -> Line<'static> {
        let m = &self.filter_matches[i];
        let item = &self.items[self.filter_item_indexes[m.position]];
        let marker = if i == self.filter_cursor {
            Span::styled("› ", ACCENT)
        } else {
            Span::raw("  ")
        };
        let mut spans = vec![
            marker,
            Span::raw("  ".repeat(item.indent_level)),
            Span::styled(format!("{} ", check_box(item)), check_style(item)),
        ];
        spans.extend(highlight(&item.text, &m.indices, item_style(item)));
        Line::from(spans)
    }

    /// Keymap rows visible inside the help popup.
    pub(super) fn help_rows(&self) -> usize {
        let fit = self
            .height
            .saturating_sub(HEADER_LINES + FOOTER_LINES + HELP_CHROME_Y)
            .max(1);
        self.help_lines().len().min(fit)
    }

    /// The keymap, in two columns when the terminal is wide enough.
    pub(super) fn help_lines(&self) -> Vec<Line<'static>> {
        if self.width < 2 * HELP_COLUMN + 4 + HELP_CHROME_X {
            return keymap_lines(KEYMAP);
        }
        let left: Vec<_> = KEYMAP.iter().step_by(2).copied().collect();
        let right: Vec<_> = KEYMAP.iter().skip(1).step_by(2).copied().collect();
        let (left, right) = (keymap_lines(&left), keymap_lines(&right));

        let rows = left.len().max(right.len());
        (0..rows)
            .map(|row| {
                let mut spans = left.get(row).cloned().unwrap_or_default().spans;
                let used: usize = spans.iter().map(Span::width).sum();
                spans.push(Span::raw(" ".repeat(HELP_COLUMN + 4 - used)));
                spans.extend(right.get(row).cloned().unwrap_or_default().spans);
                Line::from(spans)
            })
            .collect()
    }
}

/// Lines for keymap groups: a heading, its keys, and a blank line between
/// groups.
fn keymap_lines(groups: &[(&str, &[Hint])]) -> Vec<Line<'static>> {
    let mut out = Vec::new();
    for (group, keys) in groups {
        if !out.is_empty() {
            out.push(Line::raw(""));
        }
        out.push(Line::from(Span::styled(group.to_string(), HEADING)));
        for (key, action) in *keys {
            out.push(Line::from(vec![
                Span::raw("  "),
                Span::styled(format!("{key:<KEY_COLUMN$}"), KEY),
                Span::raw(*action),
            ]));
        }
    }
    out
}

fn empty_state() -> Vec<Line<'static>> {
    vec![
        Line::raw("  Nothing to do in this directory."),
        Line::from(vec![
            Span::styled("  Press ", MUTED),
            Span::styled("a", KEY),
            Span::styled(" to add an item.", MUTED),
        ]),
    ]
}

/// Fits hints into `width`, dropping lower-priority ones first and always
/// keeping the pinned hints at the end.
fn hint_line(hints: &[Hint], pinned: &[Hint], width: usize) -> Line<'static> {
    let cost = |h: &Hint| h.0.width() + 1 + h.1.width() + 3;
    let budget = width.saturating_sub(pinned.iter().map(cost).sum());

    let mut used = 0;
    let shown: Vec<&Hint> = hints
        .iter()
        .take_while(|h| {
            used += cost(h);
            used <= budget
        })
        .chain(pinned)
        .collect();

    let mut spans = Vec::new();
    for (i, (key, label)) in shown.into_iter().enumerate() {
        if i > 0 {
            spans.push(Span::raw("   "));
        }
        spans.push(Span::styled(*key, KEY));
        spans.push(Span::styled(format!(" {label}"), MUTED));
    }
    Line::from(spans)
}

/// Splits text into spans, styling the characters at `indices` as matches.
fn highlight(text: &str, indices: &[usize], base: ratatui::style::Style) -> Vec<Span<'static>> {
    text.chars()
        .enumerate()
        .map(|(i, c)| {
            let style = if indices.contains(&i) {
                base.patch(MATCH)
            } else {
                base
            };
            Span::styled(c.to_string(), style)
        })
        .collect()
}

/// Shortens a path from the left with an ellipsis so it fits `width`.
fn truncate_left(path: &str, width: usize) -> String {
    if path.width() <= width {
        return path.to_string();
    }
    let chars: Vec<char> = path.chars().collect();
    let keep = width.saturating_sub(1);
    let tail: String = chars[chars.len().saturating_sub(keep)..].iter().collect();
    format!("…{tail}")
}

fn check_box(item: &Item) -> &'static str {
    if item.checked { "[x]" } else { "[ ]" }
}

fn check_style(item: &Item) -> ratatui::style::Style {
    if item.checked { CHECKED } else { MUTED }
}

fn item_style(item: &Item) -> ratatui::style::Style {
    if item.is_active {
        ACTIVE
    } else if item.checked {
        DONE
    } else {
        ratatui::style::Style::new()
    }
}
