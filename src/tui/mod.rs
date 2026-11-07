//! Interactive terminal UI built on ratatui.

mod input;
mod theme;
mod view;

use std::io;

use fuzzy_matcher::FuzzyMatcher;
use fuzzy_matcher::skim::SkimMatcherV2;
use ratatui::DefaultTerminal;
use ratatui::crossterm::event::{
    self, DisableBracketedPaste, EnableBracketedPaste, Event, KeyCode, KeyEvent, KeyEventKind,
    KeyModifiers,
};
use ratatui::crossterm::execute;

use crate::config::home_dir;
use crate::store::{self, Item, Store};
use input::TextInput;

/// Empty rows and columns kept around the content on each side.
const FRAME_PADDING_Y: u16 = 1;
const FRAME_PADDING_X: u16 = 2;

/// Lines used by the fixed header (directory and counts, then a blank line).
const HEADER_LINES: usize = 2;
/// Lines used by the fixed footer (a blank line, then hints or a status).
const FOOTER_LINES: usize = 2;

const MAX_INDENT: usize = 5;

/// Runs the TUI until the user quits.
pub fn run(store: Store, directory: String) -> io::Result<()> {
    let mut terminal = ratatui::init();
    execute!(io::stdout(), EnableBracketedPaste)?;
    let result = run_loop(&mut terminal, App::new(store, directory));
    let _ = execute!(io::stdout(), DisableBracketedPaste);
    ratatui::restore();
    result
}

fn run_loop(terminal: &mut DefaultTerminal, mut app: App) -> io::Result<()> {
    let size = terminal.size()?;
    app.resize(size.width, size.height);
    app.reload();

    while !app.quit {
        terminal.draw(|frame| app.draw(frame))?;
        match event::read()? {
            Event::Key(key) if key.kind != KeyEventKind::Release => app.handle_key(key),
            Event::Paste(text) => app.handle_paste(&text),
            Event::Resize(width, height) => app.resize(width, height),
            _ => {}
        }
    }
    Ok(())
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum Mode {
    Normal,
    Add,
    Edit,
    Filter,
    Help,
}

/// A one-line message in the footer, cleared by the next key press.
struct Status {
    text: String,
    is_error: bool,
}

/// A filter result: the item's position among filterable items and the
/// matched character indices.
struct FilterMatch {
    position: usize,
    indices: Vec<usize>,
}

pub struct App {
    store: Store,
    directory: String,
    /// The directory as shown in the header, with the home prefix as `~`.
    display_dir: String,
    items: Vec<Item>,
    cursor: usize,
    mode: Mode,
    input: TextInput,
    /// Content size, excluding the frame padding.
    width: usize,
    height: usize,
    scroll_offset: usize,
    status: Option<Status>,
    /// First keymap row shown on the help screen.
    help_scroll: usize,
    /// In add mode, insert before the cursor instead of after.
    add_before: bool,
    /// Indent level for the new item in add mode.
    add_indent: usize,
    quit: bool,
    clipboard: Option<arboard::Clipboard>,

    filter_input: TextInput,
    /// Best match first.
    filter_matches: Vec<FilterMatch>,
    /// Maps a filterable position to its index in `items`.
    filter_item_indexes: Vec<usize>,
    filter_cursor: usize,
    filter_prev_cursor: usize,
}

impl App {
    pub fn new(store: Store, directory: String) -> App {
        App {
            store,
            display_dir: display_path(&directory),
            directory,
            items: Vec::new(),
            cursor: 0,
            mode: Mode::Normal,
            input: TextInput::new("", "New item..."),
            width: 0,
            height: 0,
            scroll_offset: 0,
            status: None,
            help_scroll: 0,
            add_before: false,
            add_indent: 0,
            quit: false,
            clipboard: None,
            filter_input: TextInput::new("/ ", ""),
            filter_matches: Vec::new(),
            filter_item_indexes: Vec::new(),
            filter_cursor: 0,
            filter_prev_cursor: 0,
        }
    }

    fn resize(&mut self, width: u16, height: u16) {
        self.width = width.saturating_sub(2 * FRAME_PADDING_X) as usize;
        self.height = height.saturating_sub(2 * FRAME_PADDING_Y) as usize;
        // 2 for the cursor, 4 for "[ ] ".
        self.input.width = self.width.saturating_sub(6);
        // 2 for the "/ " prompt.
        self.filter_input.width = self.width.saturating_sub(2);
        self.ensure_cursor_visible();
    }

    fn reload(&mut self) {
        match self.store.list(&self.directory) {
            Ok(items) => self.items = items,
            Err(e) => self.set_error(e),
        }
        if self.cursor >= self.items.len() {
            self.cursor = self.items.len().saturating_sub(1);
        }
        self.ensure_cursor_visible();
    }

    /// Records a failed store operation and reloads the items.
    fn after_change<T>(&mut self, result: store::Result<T>) {
        if let Err(e) = result {
            self.set_error(e);
        }
        self.reload();
    }

    fn set_status(&mut self, text: impl Into<String>) {
        self.status = Some(Status {
            text: text.into(),
            is_error: false,
        });
    }

    fn set_error(&mut self, err: impl std::fmt::Display) {
        self.status = Some(Status {
            text: format!("Error: {err}"),
            is_error: true,
        });
    }

    fn current(&self) -> Option<&Item> {
        self.items.get(self.cursor)
    }

    fn handle_paste(&mut self, text: &str) {
        match self.mode {
            Mode::Add | Mode::Edit => self.input.paste(text),
            Mode::Filter => {
                self.filter_input.paste(text);
                self.recompute_filter_matches();
            }
            Mode::Normal | Mode::Help => {}
        }
    }

    fn handle_key(&mut self, key: KeyEvent) {
        self.status = None;
        match self.mode {
            Mode::Add | Mode::Edit => self.update_input(key),
            Mode::Filter => self.update_filter(key),
            Mode::Help => self.update_help(key),
            Mode::Normal => self.update_normal(key),
        }
        self.ensure_cursor_visible();
    }

    fn update_normal(&mut self, key: KeyEvent) {
        match key_name(&key).as_str() {
            "q" | "esc" | "ctrl+c" => self.quit = true,
            "j" | "down" => self.cursor = (self.cursor + 1).min(self.items.len().saturating_sub(1)),
            "k" | "up" => self.cursor = self.cursor.saturating_sub(1),
            "J" => self.move_item(true),
            "K" => self.move_item(false),
            "enter" | " " => self.on_current_item(Store::toggle),
            "x" => self.on_current_item(Store::toggle_active),
            "tab" => self.on_current_item(Store::indent),
            "shift+tab" => self.on_current_item(Store::unindent),
            "a" => self.start_add(false, 0),
            "o" => self.start_add(false, self.indent_below_cursor()),
            "O" => self.start_add(true, self.current().map_or(0, |i| i.indent_level)),
            "s" => self.insert_section(),
            "e" => self.start_edit(),
            "d" => self.delete_current(),
            "c" => self.copy_current(),
            "C" => self.clean(),
            "/" => self.start_filter(),
            "?" => {
                self.mode = Mode::Help;
                self.help_scroll = 0;
            }
            _ => {}
        }
    }

    /// Applies a store operation to the item under the cursor, unless it is a
    /// section.
    fn on_current_item(&mut self, op: fn(&Store, &str, i64) -> store::Result<()>) {
        let Some(item) = self.current().filter(|i| !i.is_section) else {
            return;
        };
        let result = op(&self.store, &self.directory, item.id as i64);
        self.after_change(result);
    }

    /// Swaps the cursor item with its neighbour and follows it.
    fn move_item(&mut self, down: bool) {
        let target = if down {
            self.cursor + 1
        } else {
            self.cursor.wrapping_sub(1)
        };
        let (Some(item), Some(other)) = (self.items.get(self.cursor), self.items.get(target))
        else {
            return;
        };
        let result = self
            .store
            .swap(&self.directory, item.id as i64, other.id as i64);
        self.cursor = target;
        self.after_change(result);
    }

    /// Indent for a new item below the cursor: the deeper of the cursor item
    /// and the item after it.
    fn indent_below_cursor(&self) -> usize {
        let here = self.current().map_or(0, |i| i.indent_level);
        let next = self
            .items
            .get(self.cursor + 1)
            .map_or(0, |i| i.indent_level);
        here.max(next)
    }

    fn insert_section(&mut self) {
        let result = match self.current().map(|i| i.id) {
            Some(id) => {
                self.cursor += 1;
                self.store.insert_section_after(&self.directory, id, "")
            }
            None => self.store.add_section(&self.directory, ""),
        };
        self.after_change(result);
    }

    fn start_edit(&mut self) {
        let Some(item) = self.current().cloned() else {
            return;
        };
        self.mode = Mode::Edit;
        if item.is_section {
            self.input.placeholder = "Section title...".into();
            // Cursor (2) + "── " (3) = 5 cells before the input.
            self.input.width = self.width.saturating_sub(5);
        } else {
            self.input.placeholder = "Edit item...".into();
            self.input.width = self.width.saturating_sub(6 + item.indent_level * 2);
        }
        self.input.set_value(&item.text);
    }

    fn update_help(&mut self, key: KeyEvent) {
        let max_scroll = self.help_lines().len() - self.help_rows();
        match key_name(&key).as_str() {
            "?" | "esc" | "q" | "ctrl+c" => self.mode = Mode::Normal,
            "j" | "down" => self.help_scroll = (self.help_scroll + 1).min(max_scroll),
            "k" | "up" => self.help_scroll = self.help_scroll.saturating_sub(1),
            _ => {}
        }
    }

    fn clean(&mut self) {
        match self.store.clean(&self.directory) {
            Ok(0) => self.set_status("No done items to remove"),
            Ok(1) => self.set_status("Removed 1 done item"),
            Ok(n) => self.set_status(format!("Removed {n} done items")),
            Err(e) => self.set_error(e),
        }
        self.reload();
    }

    fn delete_current(&mut self) {
        let Some(item) = self.current().cloned() else {
            return;
        };
        let result = self.store.delete(&self.directory, item.id as i64);
        if result.is_ok() {
            let what = if item.is_section { "section" } else { "item" };
            self.set_status(format!("Deleted {what}"));
        }
        self.after_change(result);
    }

    fn start_filter(&mut self) {
        self.mode = Mode::Filter;
        self.filter_prev_cursor = self.cursor;
        self.filter_input.reset();
        self.scroll_offset = 0;
        self.recompute_filter_matches();
    }

    fn start_add(&mut self, before: bool, indent: usize) {
        self.mode = Mode::Add;
        self.add_before = before;
        self.add_indent = indent;
        self.input.reset();
        self.input.placeholder = "New item...".into();
        self.input.width = self.width.saturating_sub(6 + indent * 2);
    }

    fn copy_current(&mut self) {
        let Some(text) = self.current().map(|i| i.text.clone()) else {
            return;
        };
        if self.clipboard.is_none() {
            self.clipboard = arboard::Clipboard::new().ok();
        }
        let copied = match &mut self.clipboard {
            Some(clipboard) => clipboard.set_text(text).is_ok(),
            None => false,
        };
        if copied {
            self.set_status("Copied to clipboard");
        } else {
            self.set_error("clipboard unavailable");
        }
    }

    fn update_input(&mut self, key: KeyEvent) {
        match key_name(&key).as_str() {
            "enter" | "esc" => self.save_input(),
            "ctrl+c" => self.mode = Mode::Normal,
            "tab" if self.mode == Mode::Add => self.set_add_indent(self.add_indent + 1),
            "shift+tab" if self.mode == Mode::Add => {
                self.set_add_indent(self.add_indent.saturating_sub(1))
            }
            "tab" | "shift+tab" => {}
            _ => {
                self.input.handle_key(key);
            }
        }
    }

    fn set_add_indent(&mut self, indent: usize) {
        self.add_indent = indent.min(MAX_INDENT);
        self.input.width = self.width.saturating_sub(6 + self.add_indent * 2);
    }

    /// Saves the add or edit input and returns to normal mode. Empty input
    /// changes nothing.
    fn save_input(&mut self) {
        let text = self.input.value().trim().to_string();
        let mode = self.mode;
        self.mode = Mode::Normal;
        if text.is_empty() {
            return self.reload();
        }

        let dir = &self.directory;
        let result = match (mode, self.current().map(|i| i.id)) {
            (Mode::Edit, Some(id)) => self.store.edit(dir, id as i64, &text),
            (Mode::Edit, None) => Ok(()),
            (_, Some(id)) if self.add_before => self
                .store
                .insert_item_before(dir, &text, id, self.add_indent)
                .map(drop),
            (_, Some(id)) => {
                let result = self
                    .store
                    .insert_item_after(dir, &text, id, self.add_indent);
                self.cursor += 1;
                result.map(drop)
            }
            (_, None) => self.store.add(dir, &text).map(drop),
        };
        self.after_change(result);
    }

    fn update_filter(&mut self, key: KeyEvent) {
        match key_name(&key).as_str() {
            "enter" => {
                self.cursor = match self.filter_matches.get(self.filter_cursor) {
                    Some(m) => self.filter_item_indexes[m.position],
                    None => self.filter_prev_cursor,
                };
                self.mode = Mode::Normal;
            }
            "esc" => {
                self.cursor = self.filter_prev_cursor;
                self.mode = Mode::Normal;
            }
            "ctrl+n" | "down" => {
                if self.filter_cursor + 1 < self.filter_matches.len() {
                    self.filter_cursor += 1;
                }
            }
            "ctrl+p" | "up" => self.filter_cursor = self.filter_cursor.saturating_sub(1),
            _ => {
                if self.filter_input.handle_key(key) {
                    self.recompute_filter_matches();
                }
            }
        }
    }

    /// Fuzzy matches the filter query against non-section items, best first.
    fn recompute_filter_matches(&mut self) {
        self.filter_item_indexes = (0..self.items.len())
            .filter(|&i| !self.items[i].is_section)
            .collect();
        let query = self.filter_input.value();
        let matcher = SkimMatcherV2::default().ignore_case();
        let mut scored: Vec<(i64, FilterMatch)> = self
            .filter_item_indexes
            .iter()
            .enumerate()
            .filter_map(|(position, &i)| {
                if query.is_empty() {
                    return Some((
                        0,
                        FilterMatch {
                            position,
                            indices: Vec::new(),
                        },
                    ));
                }
                let (score, indices) = matcher.fuzzy_indices(&self.items[i].text, &query)?;
                Some((score, FilterMatch { position, indices }))
            })
            .collect();
        scored.sort_by_key(|(score, _)| std::cmp::Reverse(*score));
        self.filter_matches = scored.into_iter().map(|(_, m)| m).collect();
        self.filter_cursor = 0;
    }

    /// Index of the first non-active item when active items precede it.
    fn active_divider_index(&self) -> Option<usize> {
        if !self.items.first()?.is_active {
            return None;
        }
        self.items.iter().position(|i| !i.is_active)
    }

    /// Terminal lines the item at `i` occupies, including the add-mode input
    /// line and the gap after active items.
    fn item_line_count(&self, i: usize) -> usize {
        let item = &self.items[i];
        let mut lines = 1;
        let editing = self.mode == Mode::Edit && i == self.cursor;
        if !item.is_section && !editing && self.width > 0 {
            // Cursor (2) + indent + "[ ] " (4).
            let prefix_width = 6 + item.indent_level * 2;
            if prefix_width < self.width {
                lines = wrap(&item.text, self.width - prefix_width).len();
            }
        }
        if self.active_divider_index() == Some(i) && !item.is_section {
            lines += 1;
        }
        if self.mode == Mode::Add && i == self.cursor {
            lines += 1;
        }
        lines
    }

    /// Adjusts `scroll_offset` so the cursor item is fully visible.
    fn ensure_cursor_visible(&mut self) {
        if self.mode == Mode::Filter || self.height == 0 || self.cursor >= self.items.len() {
            return;
        }
        let available = self.height.saturating_sub(HEADER_LINES + FOOTER_LINES);
        if available == 0 {
            return;
        }

        let cursor_start: usize = (0..self.cursor).map(|i| self.item_line_count(i)).sum();
        let cursor_end = cursor_start + self.item_line_count(self.cursor);

        if cursor_start < self.scroll_offset {
            self.scroll_offset = cursor_start;
        }
        if cursor_end > self.scroll_offset + available {
            self.scroll_offset = cursor_end - available;
        }

        // Avoid blank space at the bottom after a resize or deletion.
        let total: usize = cursor_end
            + (self.cursor + 1..self.items.len())
                .map(|i| self.item_line_count(i))
                .sum::<usize>();
        self.scroll_offset = self.scroll_offset.min(total.saturating_sub(available));
    }
}

/// Returns the directory with the home prefix shown as `~`.
fn display_path(dir: &str) -> String {
    let Some(home) = home_dir().map(|h| h.to_string_lossy().into_owned()) else {
        return dir.to_string();
    };
    if dir == home {
        return "~".to_string();
    }
    match dir.strip_prefix(&format!("{home}/")) {
        Some(rest) => format!("~/{rest}"),
        None => dir.to_string(),
    }
}

/// Word-wraps text to `width` cells, breaking words longer than a line.
fn wrap(text: &str, width: usize) -> Vec<String> {
    let lines: Vec<String> = textwrap::wrap(text, width.max(1))
        .into_iter()
        .map(|l| l.into_owned())
        .collect();
    if lines.is_empty() {
        vec![String::new()]
    } else {
        lines
    }
}

/// Names a key the way the keymap refers to it, e.g. "j", "J", "ctrl+c",
/// "shift+tab", "enter".
fn key_name(key: &KeyEvent) -> String {
    let ctrl = key.modifiers.contains(KeyModifiers::CONTROL);
    let alt = key.modifiers.contains(KeyModifiers::ALT);
    let base = match key.code {
        KeyCode::Char(c) => c.to_string(),
        KeyCode::Enter => "enter".into(),
        KeyCode::Esc => "esc".into(),
        KeyCode::Tab => "tab".into(),
        KeyCode::BackTab => "shift+tab".into(),
        KeyCode::Up => "up".into(),
        KeyCode::Down => "down".into(),
        KeyCode::Left => "left".into(),
        KeyCode::Right => "right".into(),
        KeyCode::Backspace => "backspace".into(),
        _ => return String::new(),
    };
    match (ctrl, alt) {
        (true, _) => format!("ctrl+{base}"),
        (_, true) => format!("alt+{base}"),
        _ => base,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use ratatui::Terminal;
    use ratatui::backend::TestBackend;
    use unicode_width::UnicodeWidthStr;

    const DIR: &str = "/tmp/project";

    fn app_with(items: &[&str]) -> (tempfile::TempDir, App) {
        let tmp = tempfile::tempdir().unwrap();
        let store = Store::open(tmp.path().join("todo.md")).unwrap();
        for text in items {
            store.add(DIR, text).unwrap();
        }
        let mut app = App::new(store, DIR.to_string());
        app.resize(40, 10);
        app.reload();
        (tmp, app)
    }

    fn press(app: &mut App, keys: &str) {
        for c in keys.chars() {
            app.handle_key(KeyEvent::new(KeyCode::Char(c), KeyModifiers::NONE));
        }
    }

    fn press_code(app: &mut App, code: KeyCode) {
        app.handle_key(KeyEvent::new(code, KeyModifiers::NONE));
    }

    fn texts(app: &App) -> Vec<String> {
        app.store
            .list(DIR)
            .unwrap()
            .into_iter()
            .map(|i| i.text)
            .collect()
    }

    fn screen(app: &mut App, width: u16, height: u16) -> Vec<String> {
        let mut terminal = Terminal::new(TestBackend::new(width, height)).unwrap();
        app.resize(width, height);
        terminal.draw(|f| app.draw(f)).unwrap();
        let buffer = terminal.backend().buffer();
        (0..height)
            .map(|y| {
                (0..width)
                    .map(|x| buffer[(x, y)].symbol().to_string())
                    .collect::<String>()
            })
            .collect()
    }

    #[test]
    fn view_keeps_margin_around_content() {
        let mut items =
            vec!["a long item that wraps past the right edge of the screen".to_string()];
        items.extend((0..20).map(|i| format!("item {i}")));
        let refs: Vec<&str> = items.iter().map(String::as_str).collect();
        let (_tmp, mut app) = app_with(&refs);

        let lines = screen(&mut app, 40, 10);
        assert_eq!(lines[0].trim(), "");
        assert_eq!(lines[1].trim_end(), "  /tmp/project");
        assert!(lines[2].trim().is_empty());
        assert!(
            lines[3].starts_with("  › [ ] a long item"),
            "{:?}",
            lines[3]
        );
        assert!(
            lines[4].starts_with("        "),
            "wrapped line keeps prefix indent"
        );
        assert!(lines[8].contains("? help"), "footer: {:?}", lines[8]);
        assert_eq!(lines[9].trim(), "");
        for line in &lines {
            assert!(line.trim_end().width() <= 38, "line too wide: {line:?}");
        }
    }

    #[test]
    fn narrow_footer_keeps_help_hint() {
        let (_tmp, mut app) = app_with(&["a"]);
        let lines = screen(&mut app, 30, 8);
        assert_eq!(lines[6].trim(), "a add   ? help   q quit");
    }

    #[test]
    fn actions_report_status_until_next_key() {
        let (_tmp, mut app) = app_with(&["a", "b"]);
        press(&mut app, "C");
        assert_eq!(app.status.as_ref().unwrap().text, "No done items to remove");
        press(&mut app, " jx");
        press(&mut app, "C");
        assert_eq!(app.status.as_ref().unwrap().text, "Removed 1 done item");
        let lines = screen(&mut app, 40, 10);
        assert_eq!(lines[8].trim(), "Removed 1 done item");
        press(&mut app, "j");
        assert!(app.status.is_none());
    }

    #[test]
    fn help_screen_opens_and_closes() {
        let (_tmp, mut app) = app_with(&["a"]);
        press(&mut app, "?");
        let lines = screen(&mut app, 60, 30);
        assert!(lines.iter().any(|l| l.contains("┌ Keys ")), "popup border");
        assert!(lines.iter().any(|l| l.contains("mark done or not done")));
        assert!(
            lines[1].contains("/tmp/project"),
            "list stays behind the popup"
        );
        press(&mut app, "j");
        assert_eq!(app.mode, Mode::Help, "j scrolls instead of closing");
        press(&mut app, "?");
        assert_eq!(app.mode, Mode::Normal);
    }

    #[test]
    fn add_after_cursor_inherits_indent() {
        let (_tmp, mut app) = app_with(&["a", "b"]);
        press_code(&mut app, KeyCode::Tab);
        press(&mut app, "ox");
        press_code(&mut app, KeyCode::Enter);
        assert_eq!(texts(&app), ["a", "x", "b"]);
        assert_eq!(app.store.get(DIR, 2).unwrap().indent_level, 1);
        assert_eq!(app.cursor, 1);
    }

    #[test]
    fn add_before_and_cancel() {
        let (_tmp, mut app) = app_with(&["a"]);
        press(&mut app, "Oz");
        press_code(&mut app, KeyCode::Esc);
        assert_eq!(texts(&app), ["z", "a"]);
        press(&mut app, "anope");
        app.handle_key(KeyEvent::new(KeyCode::Char('c'), KeyModifiers::CONTROL));
        assert_eq!(texts(&app), ["z", "a"]);
        assert_eq!(app.mode, Mode::Normal);
    }

    #[test]
    fn reorder_toggle_and_delete() {
        let (_tmp, mut app) = app_with(&["a", "b", "c"]);
        press(&mut app, "J");
        assert_eq!(texts(&app), ["b", "a", "c"]);
        assert_eq!(app.cursor, 1);
        press(&mut app, " ");
        assert!(app.store.get(DIR, 2).unwrap().checked);
        press(&mut app, "d");
        assert_eq!(texts(&app), ["b", "c"]);
    }

    #[test]
    fn edit_section_title() {
        let (_tmp, mut app) = app_with(&["a"]);
        press(&mut app, "sePlan");
        press_code(&mut app, KeyCode::Enter);
        let section = app.store.get(DIR, 2).unwrap();
        assert!(section.is_section);
        assert_eq!(section.text, "Plan");
    }

    #[test]
    fn filter_jumps_to_selected_match() {
        let (_tmp, mut app) = app_with(&["buy milk", "fix bug", "call mom"]);
        press(&mut app, "/fxb");
        assert_eq!(app.filter_matches.len(), 1);
        assert_eq!(app.filter_matches[0].indices, [0, 2, 4]);
        press_code(&mut app, KeyCode::Enter);
        assert_eq!(app.cursor, 1);

        press(&mut app, "/");
        press_code(&mut app, KeyCode::Down);
        press_code(&mut app, KeyCode::Down);
        press_code(&mut app, KeyCode::Enter);
        assert_eq!(app.cursor, 2, "arrow keys move through matches");

        press(&mut app, "/zzz");
        press_code(&mut app, KeyCode::Enter);
        assert_eq!(app.cursor, 2, "no match keeps the previous cursor");
    }

    #[test]
    fn active_items_float_above_a_gap() {
        let (_tmp, mut app) = app_with(&["a", "b"]);
        press(&mut app, "jx");
        let lines = screen(&mut app, 40, 10);
        assert!(lines[3].contains("[ ] b"), "{:?}", lines[3]);
        assert_eq!(lines[4].trim(), "");
        assert!(lines[5].contains("[ ] a"), "{:?}", lines[5]);
    }
}
