//! Single-line text input with readline-style editing keys.

use ratatui::crossterm::event::{KeyCode, KeyEvent, KeyModifiers};
use ratatui::text::Span;
use unicode_width::UnicodeWidthChar;

const CHAR_LIMIT: usize = 256;

#[derive(Debug, Default)]
pub struct TextInput {
    value: Vec<char>,
    pos: usize,
    /// First visible character when the value is wider than `width`.
    offset: usize,
    /// Maximum cells used to display the value; 0 means unlimited.
    pub width: usize,
    pub placeholder: String,
    pub prompt: String,
}

impl TextInput {
    pub fn new(prompt: &str, placeholder: &str) -> TextInput {
        TextInput {
            prompt: prompt.to_string(),
            placeholder: placeholder.to_string(),
            ..Default::default()
        }
    }

    pub fn value(&self) -> String {
        self.value.iter().collect()
    }

    pub fn reset(&mut self) {
        self.value.clear();
        self.pos = 0;
        self.offset = 0;
    }

    /// Replaces the value and moves the cursor to the end.
    pub fn set_value(&mut self, value: &str) {
        self.value = value.chars().take(CHAR_LIMIT).collect();
        self.pos = self.value.len();
        self.offset = 0;
    }

    /// Inserts pasted text, turning tabs and newlines into spaces.
    pub fn paste(&mut self, text: &str) {
        for c in text.chars() {
            let c = if matches!(c, '\n' | '\r' | '\t') {
                ' '
            } else {
                c
            };
            self.insert(c);
        }
    }

    fn insert(&mut self, c: char) {
        if self.value.len() >= CHAR_LIMIT || c.is_control() {
            return;
        }
        self.value.insert(self.pos, c);
        self.pos += 1;
    }

    /// Applies an editing key. Returns false when the key is not an editing
    /// key.
    pub fn handle_key(&mut self, key: KeyEvent) -> bool {
        let ctrl = key.modifiers.contains(KeyModifiers::CONTROL);
        let alt = key.modifiers.contains(KeyModifiers::ALT);
        match key.code {
            KeyCode::Char('a') if ctrl => self.pos = 0,
            KeyCode::Char('e') if ctrl => self.pos = self.value.len(),
            KeyCode::Char('b') if ctrl => self.pos = self.pos.saturating_sub(1),
            KeyCode::Char('f') if ctrl => self.pos = (self.pos + 1).min(self.value.len()),
            KeyCode::Char('h') if ctrl => self.delete_back(),
            KeyCode::Char('d') if ctrl => self.delete_forward(),
            KeyCode::Char('u') if ctrl => {
                self.value.drain(..self.pos);
                self.pos = 0;
            }
            KeyCode::Char('k') if ctrl => self.value.truncate(self.pos),
            KeyCode::Char('w') if ctrl => self.delete_word_back(),
            KeyCode::Char('b') if alt => self.pos = self.word_start(),
            KeyCode::Char('f') if alt => self.pos = self.word_end(),
            KeyCode::Char('d') if alt => {
                let end = self.word_end();
                self.value.drain(self.pos..end);
            }
            KeyCode::Char(_) if ctrl || alt => return false,
            KeyCode::Char(c) => self.insert(c),
            KeyCode::Backspace if alt || ctrl => self.delete_word_back(),
            KeyCode::Backspace => self.delete_back(),
            KeyCode::Delete => self.delete_forward(),
            KeyCode::Left if alt || ctrl => self.pos = self.word_start(),
            KeyCode::Right if alt || ctrl => self.pos = self.word_end(),
            KeyCode::Left => self.pos = self.pos.saturating_sub(1),
            KeyCode::Right => self.pos = (self.pos + 1).min(self.value.len()),
            KeyCode::Home => self.pos = 0,
            KeyCode::End => self.pos = self.value.len(),
            _ => return false,
        }
        true
    }

    fn delete_back(&mut self) {
        if self.pos > 0 {
            self.pos -= 1;
            self.value.remove(self.pos);
        }
    }

    fn delete_forward(&mut self) {
        if self.pos < self.value.len() {
            self.value.remove(self.pos);
        }
    }

    fn delete_word_back(&mut self) {
        let start = self.word_start();
        self.value.drain(start..self.pos);
        self.pos = start;
    }

    fn word_start(&self) -> usize {
        let mut i = self.pos;
        while i > 0 && self.value[i - 1].is_whitespace() {
            i -= 1;
        }
        while i > 0 && !self.value[i - 1].is_whitespace() {
            i -= 1;
        }
        i
    }

    fn word_end(&self) -> usize {
        let mut i = self.pos;
        while i < self.value.len() && self.value[i].is_whitespace() {
            i += 1;
        }
        while i < self.value.len() && !self.value[i].is_whitespace() {
            i += 1;
        }
        i
    }

    /// Renders the prompt and the visible part of the value. Returns the
    /// spans and the cursor column relative to the start of the prompt.
    pub fn render(&mut self) -> (Vec<Span<'static>>, usize) {
        let prompt_width = self.prompt.chars().count();
        let mut spans = vec![Span::raw(self.prompt.clone())];

        if self.value.is_empty() {
            if !self.placeholder.is_empty() {
                spans.push(Span::styled(self.placeholder.clone(), super::theme::MUTED));
            }
            return (spans, prompt_width);
        }

        self.scroll_to_cursor();
        let mut used = 0;
        let mut visible = String::new();
        let mut cursor_col = None;
        for (i, &c) in self.value.iter().enumerate().skip(self.offset) {
            if i == self.pos {
                cursor_col = Some(used);
            }
            let w = c.width().unwrap_or(0);
            if self.width > 0 && used + w > self.width {
                break;
            }
            visible.push(c);
            used += w;
        }
        spans.push(Span::raw(visible));
        (spans, prompt_width + cursor_col.unwrap_or(used))
    }

    /// Adjusts `offset` so the cursor, plus one cell for it at the end of the
    /// value, fits inside `width`.
    fn scroll_to_cursor(&mut self) {
        if self.width == 0 {
            self.offset = 0;
            return;
        }
        if self.pos < self.offset {
            self.offset = self.pos;
        }
        let cells = |from: usize, to: usize| -> usize {
            self.value[from..to]
                .iter()
                .map(|c| c.width().unwrap_or(0))
                .sum()
        };
        let cursor_cell = if self.pos == self.value.len() { 1 } else { 0 };
        while self.offset < self.pos && cells(self.offset, self.pos) + cursor_cell > self.width {
            self.offset += 1;
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn key(code: KeyCode, modifiers: KeyModifiers) -> KeyEvent {
        KeyEvent::new(code, modifiers)
    }

    fn type_str(input: &mut TextInput, s: &str) {
        for c in s.chars() {
            input.handle_key(key(KeyCode::Char(c), KeyModifiers::NONE));
        }
    }

    #[test]
    fn edits_at_cursor() {
        let mut input = TextInput::new("", "");
        type_str(&mut input, "hello world");
        input.handle_key(key(KeyCode::Char('w'), KeyModifiers::CONTROL));
        assert_eq!(input.value(), "hello ");
        input.handle_key(key(KeyCode::Home, KeyModifiers::NONE));
        type_str(&mut input, ">");
        assert_eq!(input.value(), ">hello ");
        input.handle_key(key(KeyCode::Char('k'), KeyModifiers::CONTROL));
        assert_eq!(input.value(), ">");
    }

    #[test]
    fn scrolls_long_values_to_keep_cursor_visible() {
        let mut input = TextInput::new("", "");
        input.width = 5;
        type_str(&mut input, "abcdefgh");
        let (spans, col) = input.render();
        assert_eq!(spans[1].content, "efgh");
        assert_eq!(col, 4);
    }
}
