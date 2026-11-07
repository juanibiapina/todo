//! Markdown-backed storage for todo items, scoped per directory.
//!
//! Items are addressed by their 1-based position within a directory. Every
//! operation reads the file, applies the change, and writes it back
//! atomically.

use std::fmt;
use std::fs;
use std::io::{self, Write};
use std::path::{Path, PathBuf};

use crate::config::home_dir;

/// A single todo item or section divider.
#[derive(Debug, Clone, PartialEq, Eq, Default)]
pub struct Item {
    pub id: usize,
    pub text: String,
    pub checked: bool,
    pub is_section: bool,
    pub is_active: bool,
    pub indent_level: usize,
}

const MAX_INDENT: usize = 5;

#[derive(Debug)]
pub enum Error {
    NotFound(i64),
    IsSection(i64),
    Io(String, io::Error),
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Error::NotFound(id) => write!(f, "item {id} not found"),
            Error::IsSection(id) => write!(f, "item {id} is a section"),
            Error::Io(context, err) if context.is_empty() => write!(f, "{err}"),
            Error::Io(context, err) => write!(f, "{context}: {err}"),
        }
    }
}

impl std::error::Error for Error {}

fn io_err(context: &str) -> impl FnOnce(io::Error) -> Error + '_ {
    move |e| Error::Io(context.to_string(), e)
}

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Default)]
struct FileData {
    directories: Vec<DirectoryData>,
}

#[derive(Debug, Default)]
struct DirectoryData {
    /// Expanded (absolute) path.
    path: String,
    /// "YYYY-MM-DD", or empty when unset.
    last_clean: String,
    items: Vec<FileItem>,
}

#[derive(Debug, Default, Clone)]
struct FileItem {
    text: String,
    checked: bool,
    is_section: bool,
    is_active: bool,
    indent_level: usize,
}

impl FileItem {
    fn to_item(&self, id: usize) -> Item {
        Item {
            id,
            text: self.text.clone(),
            checked: self.checked,
            is_section: self.is_section,
            is_active: self.is_active,
            indent_level: self.indent_level,
        }
    }
}

impl FileData {
    fn find(&mut self, path: &str) -> Option<&mut DirectoryData> {
        self.directories.iter_mut().find(|d| d.path == path)
    }

    fn get_or_create(&mut self, path: &str) -> &mut DirectoryData {
        let idx = match self.directories.iter().position(|d| d.path == path) {
            Some(idx) => idx,
            None => {
                self.directories.push(DirectoryData {
                    path: path.to_string(),
                    ..Default::default()
                });
                self.directories.len() - 1
            }
        };
        &mut self.directories[idx]
    }
}

/// Provides access to the todo file.
pub struct Store {
    file_path: PathBuf,
    today: Box<dyn Fn() -> String>,
}

impl Store {
    /// Creates a store backed by the given Markdown file, creating its parent
    /// directory if needed.
    pub fn open(file_path: impl Into<PathBuf>) -> Result<Store> {
        let file_path = file_path.into();
        fs::create_dir_all(parent_dir(&file_path)).map_err(io_err("creating directory"))?;
        Ok(Store {
            file_path,
            today: Box::new(|| chrono::Local::now().format("%Y-%m-%d").to_string()),
        })
    }

    /// Replaces the clock used by [`Store::clean_if_new_day`].
    #[cfg(test)]
    fn with_today(mut self, today: impl Fn() -> String + 'static) -> Store {
        self.today = Box::new(today);
        self
    }

    /// Appends a new unchecked item for the given directory.
    pub fn add(&self, directory: &str, text: &str) -> Result<Item> {
        self.append(directory, text, false)
    }

    /// Appends a new section divider for the given directory. The title may
    /// be empty.
    pub fn add_section(&self, directory: &str, title: &str) -> Result<Item> {
        self.append(directory, title, true)
    }

    fn append(&self, directory: &str, text: &str, is_section: bool) -> Result<Item> {
        let mut data = self.read()?;
        let dir = data.get_or_create(directory);
        let item = FileItem {
            text: text.to_string(),
            is_section,
            ..Default::default()
        };
        dir.items.push(item.clone());
        let id = dir.items.len();
        self.write(&mut data)?;
        Ok(item.to_item(id))
    }

    /// Inserts a new item right after the given position.
    pub fn insert_item_after(
        &self,
        directory: &str,
        text: &str,
        after_id: usize,
        indent_level: usize,
    ) -> Result<Item> {
        self.insert_at(directory, after_id, after_id, text, false, indent_level)
    }

    /// Inserts a new item right before the given position.
    pub fn insert_item_before(
        &self,
        directory: &str,
        text: &str,
        before_id: usize,
        indent_level: usize,
    ) -> Result<Item> {
        self.insert_at(
            directory,
            before_id,
            before_id - 1,
            text,
            false,
            indent_level,
        )
    }

    /// Inserts a new section divider right after the given position.
    pub fn insert_section_after(
        &self,
        directory: &str,
        after_id: usize,
        title: &str,
    ) -> Result<Item> {
        self.insert_at(directory, after_id, after_id, title, true, 0)
    }

    /// Inserts at the 0-based `index` after checking that the 1-based `anchor`
    /// exists.
    fn insert_at(
        &self,
        directory: &str,
        anchor: usize,
        index: usize,
        text: &str,
        is_section: bool,
        indent_level: usize,
    ) -> Result<Item> {
        let mut data = self.read()?;
        let dir = data.find(directory).ok_or(Error::NotFound(anchor as i64))?;
        check_id(dir, anchor as i64)?;
        let item = FileItem {
            text: text.to_string(),
            is_section,
            indent_level,
            ..Default::default()
        };
        dir.items.insert(index, item.clone());
        self.write(&mut data)?;
        Ok(item.to_item(index + 1))
    }

    /// Returns all items for the directory, active items first, then by
    /// position.
    pub fn list(&self, directory: &str) -> Result<Vec<Item>> {
        let mut data = self.read()?;
        let Some(dir) = data.find(directory) else {
            return Ok(Vec::new());
        };
        let mut items: Vec<Item> = dir
            .items
            .iter()
            .enumerate()
            .map(|(i, fi)| fi.to_item(i + 1))
            .collect();
        items.sort_by_key(|item| !item.is_active);
        Ok(items)
    }

    /// Returns a single item by 1-based position.
    pub fn get(&self, directory: &str, id: i64) -> Result<Item> {
        let mut data = self.read()?;
        let dir = data.find(directory).ok_or(Error::NotFound(id))?;
        let idx = check_id(dir, id)?;
        Ok(dir.items[idx].to_item(id as usize))
    }

    /// Marks an item as checked, clearing its active state.
    pub fn check(&self, directory: &str, id: i64) -> Result<()> {
        self.modify_item(directory, id, |fi| {
            fi.checked = true;
            fi.is_active = false;
        })
    }

    /// Marks an item as unchecked.
    pub fn uncheck(&self, directory: &str, id: i64) -> Result<()> {
        self.modify_item(directory, id, |fi| fi.checked = false)
    }

    /// Flips the checked state of an item. Checking clears its active state.
    pub fn toggle(&self, directory: &str, id: i64) -> Result<()> {
        self.modify_item(directory, id, |fi| {
            fi.checked = !fi.checked;
            if fi.checked {
                fi.is_active = false;
            }
        })
    }

    /// Flips the active state of an item.
    pub fn toggle_active(&self, directory: &str, id: i64) -> Result<()> {
        self.modify_item(directory, id, |fi| fi.is_active = !fi.is_active)
    }

    /// Increments the indent level of an item (max 5).
    pub fn indent(&self, directory: &str, id: i64) -> Result<()> {
        self.modify_item(directory, id, |fi| {
            fi.indent_level = (fi.indent_level + 1).min(MAX_INDENT)
        })
    }

    /// Decrements the indent level of an item (min 0).
    pub fn unindent(&self, directory: &str, id: i64) -> Result<()> {
        self.modify_item(directory, id, |fi| {
            fi.indent_level = fi.indent_level.saturating_sub(1)
        })
    }

    /// Applies `change` to a non-section item.
    fn modify_item(
        &self,
        directory: &str,
        id: i64,
        change: impl FnOnce(&mut FileItem),
    ) -> Result<()> {
        let mut data = self.read()?;
        let dir = data.find(directory).ok_or(Error::NotFound(id))?;
        let idx = check_id(dir, id)?;
        let fi = &mut dir.items[idx];
        if fi.is_section {
            return Err(Error::IsSection(id));
        }
        change(fi);
        self.write(&mut data)
    }

    /// Updates the text of an item. For sections this updates the title.
    pub fn edit(&self, directory: &str, id: i64, text: &str) -> Result<()> {
        let mut data = self.read()?;
        let dir = data.find(directory).ok_or(Error::NotFound(id))?;
        let idx = check_id(dir, id)?;
        dir.items[idx].text = text.to_string();
        self.write(&mut data)
    }

    /// Removes a single item by position.
    pub fn delete(&self, directory: &str, id: i64) -> Result<()> {
        let mut data = self.read()?;
        let dir = data.find(directory).ok_or(Error::NotFound(id))?;
        let idx = check_id(dir, id)?;
        dir.items.remove(idx);
        self.write(&mut data)
    }

    /// Exchanges two items within the same directory.
    pub fn swap(&self, directory: &str, id1: i64, id2: i64) -> Result<()> {
        let mut data = self.read()?;
        let dir = data.find(directory).ok_or(Error::NotFound(id1))?;
        let a = check_id(dir, id1)?;
        let b = check_id(dir, id2)?;
        dir.items.swap(a, b);
        self.write(&mut data)
    }

    /// Deletes all checked items for the directory. Returns how many were
    /// deleted.
    pub fn clean(&self, directory: &str) -> Result<usize> {
        let mut data = self.read()?;
        let Some(dir) = data.find(directory) else {
            return Ok(0);
        };
        let removed = remove_checked(dir);
        self.write(&mut data)?;
        Ok(removed)
    }

    /// Cleans checked items when a new day begins. On first use for a
    /// directory it records today's date without cleaning.
    pub fn clean_if_new_day(&self, directory: &str) -> Result<()> {
        let mut data = self.read()?;
        let today = (self.today)();
        let Some(dir) = data.find(directory) else {
            return Ok(());
        };

        if dir.last_clean == today {
            return Ok(());
        }
        if !dir.last_clean.is_empty() {
            remove_checked(dir);
        }
        dir.last_clean = today;
        self.write(&mut data)
    }

    fn read(&self) -> Result<FileData> {
        match fs::read_to_string(&self.file_path) {
            Ok(content) => Ok(parse_markdown(&content, home_str().as_deref())),
            Err(e) if e.kind() == io::ErrorKind::NotFound => Ok(FileData::default()),
            Err(e) => Err(io_err("")(e)),
        }
    }

    fn write(&self, data: &mut FileData) -> Result<()> {
        data.directories.retain(|d| !d.items.is_empty());
        let content = render_markdown(data, home_str().as_deref());

        let mut tmp = tempfile::Builder::new()
            .prefix(".todo-")
            .suffix(".tmp")
            .tempfile_in(parent_dir(&self.file_path))
            .map_err(io_err("creating temp file"))?;
        tmp.write_all(content.as_bytes())
            .map_err(io_err("writing temp file"))?;
        tmp.persist(&self.file_path)
            .map_err(|e| io_err("renaming temp file")(e.error))?;
        Ok(())
    }
}

/// Validates a 1-based id and returns the 0-based index.
fn check_id(dir: &DirectoryData, id: i64) -> Result<usize> {
    if id < 1 || id as usize > dir.items.len() {
        return Err(Error::NotFound(id));
    }
    Ok(id as usize - 1)
}

fn remove_checked(dir: &mut DirectoryData) -> usize {
    let before = dir.items.len();
    dir.items.retain(|fi| !fi.checked || fi.is_section);
    before - dir.items.len()
}

fn parent_dir(path: &Path) -> &Path {
    match path.parent() {
        Some(p) if !p.as_os_str().is_empty() => p,
        _ => Path::new("."),
    }
}

fn home_str() -> Option<String> {
    home_dir().map(|h| h.to_string_lossy().into_owned())
}

fn parse_markdown(content: &str, home: Option<&str>) -> FileData {
    let mut data = FileData::default();

    for line in content.split('\n') {
        let trimmed = line.trim();

        // Top-level heading: not data.
        if trimmed.starts_with("# ") {
            continue;
        }

        // Directory heading.
        if let Some(path) = trimmed.strip_prefix("## ") {
            let path = match (path, home) {
                ("~", Some(home)) => home.to_string(),
                (p, Some(home)) if p.starts_with("~/") => {
                    Path::new(home).join(&p[2..]).to_string_lossy().into_owned()
                }
                (p, _) => p.to_string(),
            };
            data.directories.push(DirectoryData {
                path,
                ..Default::default()
            });
            continue;
        }

        let Some(current) = data.directories.last_mut() else {
            continue;
        };

        if let Some(date) = trimmed
            .strip_prefix("<!-- last-clean: ")
            .and_then(|s| s.strip_suffix(" -->"))
        {
            current.last_clean = date.to_string();
            continue;
        }

        // Section divider, optionally with a title.
        if trimmed == "---" {
            current.items.push(FileItem {
                is_section: true,
                ..Default::default()
            });
            continue;
        }
        if let Some(title) = trimmed.strip_prefix("--- ") {
            current.items.push(FileItem {
                is_section: true,
                text: title.trim().to_string(),
                ..Default::default()
            });
            continue;
        }

        let (checked, text) = if let Some(text) = trimmed.strip_prefix("- [ ] ") {
            (false, text)
        } else if let Some(text) = trimmed.strip_prefix("- [x] ") {
            (true, text)
        } else {
            continue;
        };
        let (text, is_active) = match text.strip_suffix(" @active") {
            Some(text) => (text, true),
            None => (text, false),
        };
        let leading_spaces = line.len() - line.trim_start_matches(' ').len();
        current.items.push(FileItem {
            text: text.to_string(),
            checked,
            is_active,
            indent_level: leading_spaces / 2,
            is_section: false,
        });
    }

    data
}

fn render_markdown(data: &FileData, home: Option<&str>) -> String {
    let mut out = String::from("# TODO\n\n");

    for (i, dir) in data.directories.iter().enumerate() {
        if i > 0 {
            out.push('\n');
        }
        out.push_str(&format!("## {}\n\n", collapse_home(&dir.path, home)));
        if !dir.last_clean.is_empty() {
            out.push_str(&format!("<!-- last-clean: {} -->\n", dir.last_clean));
        }
        for item in &dir.items {
            out.push_str(&render_item(item));
        }
    }

    out
}

/// Replaces the home directory prefix with `~`.
fn collapse_home(path: &str, home: Option<&str>) -> String {
    let Some(home) = home else {
        return path.to_string();
    };
    if path == home {
        return "~".to_string();
    }
    match path.strip_prefix(&format!("{home}/")) {
        Some(rest) => format!("~/{rest}"),
        None => path.to_string(),
    }
}

/// Renders one item as a Markdown line, including the trailing newline.
fn render_item(item: &FileItem) -> String {
    if item.is_section && item.text.is_empty() {
        return "---\n".to_string();
    }
    if item.is_section {
        return format!("--- {}\n", item.text);
    }
    let check = if item.checked { "[x]" } else { "[ ]" };
    let active = if item.is_active { " @active" } else { "" };
    format!(
        "{}- {check} {}{active}\n",
        "  ".repeat(item.indent_level),
        item.text
    )
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::cell::RefCell;
    use std::rc::Rc;

    const DIR: &str = "/tmp/project";
    const OTHER: &str = "/tmp/other";

    fn new_store() -> (tempfile::TempDir, Store) {
        let tmp = tempfile::tempdir().unwrap();
        let store = Store::open(tmp.path().join("todo.md")).unwrap();
        (tmp, store)
    }

    fn texts(store: &Store, dir: &str) -> Vec<String> {
        store
            .list(dir)
            .unwrap()
            .into_iter()
            .map(|i| i.text)
            .collect()
    }

    fn file(store: &Store) -> String {
        fs::read_to_string(&store.file_path).unwrap()
    }

    #[test]
    fn add_assigns_sequential_ids_per_directory() {
        let (_tmp, s) = new_store();
        assert_eq!(s.add(DIR, "a").unwrap().id, 1);
        assert_eq!(s.add(DIR, "b").unwrap().id, 2);
        assert_eq!(s.add(OTHER, "c").unwrap().id, 1);
        assert_eq!(texts(&s, DIR), ["a", "b"]);
        assert_eq!(texts(&s, OTHER), ["c"]);
        assert!(s.list("/nowhere").unwrap().is_empty());
    }

    #[test]
    fn open_creates_parent_directory() {
        let tmp = tempfile::tempdir().unwrap();
        let path = tmp.path().join("a/b/todo.md");
        let s = Store::open(&path).unwrap();
        s.add(DIR, "x").unwrap();
        assert!(path.exists());
    }

    #[test]
    fn missing_ids_and_directories_are_not_found() {
        let (_tmp, s) = new_store();
        s.add(DIR, "a").unwrap();
        for id in [0, 2, -1] {
            assert_eq!(
                s.check(DIR, id).unwrap_err().to_string(),
                format!("item {id} not found")
            );
        }
        assert!(matches!(s.edit(OTHER, 1, "x"), Err(Error::NotFound(1))));
        assert!(matches!(s.delete(OTHER, 1), Err(Error::NotFound(1))));
        assert!(matches!(s.swap(DIR, 1, 5), Err(Error::NotFound(5))));
        assert!(matches!(s.get(DIR, 3), Err(Error::NotFound(3))));
    }

    #[test]
    fn check_uncheck_toggle() {
        let (_tmp, s) = new_store();
        s.add(DIR, "a").unwrap();
        s.check(DIR, 1).unwrap();
        s.check(DIR, 1).unwrap();
        assert!(s.get(DIR, 1).unwrap().checked);
        s.uncheck(DIR, 1).unwrap();
        assert!(!s.get(DIR, 1).unwrap().checked);
        s.toggle(DIR, 1).unwrap();
        assert!(s.get(DIR, 1).unwrap().checked);
    }

    #[test]
    fn sections_reject_item_operations_but_can_be_renamed() {
        let (_tmp, s) = new_store();
        s.add_section(DIR, "").unwrap();
        for result in [
            s.check(DIR, 1),
            s.uncheck(DIR, 1),
            s.toggle(DIR, 1),
            s.toggle_active(DIR, 1),
            s.indent(DIR, 1),
            s.unindent(DIR, 1),
        ] {
            assert_eq!(result.unwrap_err().to_string(), "item 1 is a section");
        }
        s.edit(DIR, 1, "Planning").unwrap();
        let item = s.get(DIR, 1).unwrap();
        assert!(item.is_section);
        assert_eq!(item.text, "Planning");
    }

    #[test]
    fn edit_preserves_checked_state() {
        let (_tmp, s) = new_store();
        s.add(DIR, "a").unwrap();
        s.check(DIR, 1).unwrap();
        s.edit(DIR, 1, "b").unwrap();
        let item = s.get(DIR, 1).unwrap();
        assert_eq!((item.text.as_str(), item.checked), ("b", true));
    }

    #[test]
    fn delete_and_swap_reorder_positions() {
        let (_tmp, s) = new_store();
        for t in ["a", "b", "c"] {
            s.add(DIR, t).unwrap();
        }
        s.swap(DIR, 1, 3).unwrap();
        assert_eq!(texts(&s, DIR), ["c", "b", "a"]);
        s.delete(DIR, 2).unwrap();
        assert_eq!(texts(&s, DIR), ["c", "a"]);
        assert_eq!(s.get(DIR, 2).unwrap().text, "a");
    }

    #[test]
    fn deleting_last_item_removes_directory_heading() {
        let (_tmp, s) = new_store();
        s.add(DIR, "a").unwrap();
        s.delete(DIR, 1).unwrap();
        assert_eq!(file(&s), "# TODO\n\n");
    }

    #[test]
    fn clean_removes_checked_items_only_in_directory() {
        let (_tmp, s) = new_store();
        s.add(DIR, "keep").unwrap();
        s.add(DIR, "drop").unwrap();
        s.add_section(DIR, "").unwrap();
        s.add(OTHER, "other").unwrap();
        s.check(DIR, 2).unwrap();
        s.check(OTHER, 1).unwrap();
        assert_eq!(s.clean(DIR).unwrap(), 1);
        assert_eq!(texts(&s, DIR), ["keep", ""]);
        assert_eq!(texts(&s, OTHER), ["other"]);
        assert_eq!(s.clean("/nowhere").unwrap(), 0);
    }

    #[test]
    fn inserts_items_and_sections() {
        let (_tmp, s) = new_store();
        s.add(DIR, "a").unwrap();
        s.add(DIR, "c").unwrap();
        let b = s.insert_item_after(DIR, "b", 1, 2).unwrap();
        assert_eq!((b.id, b.indent_level), (2, 2));
        let first = s.insert_item_before(DIR, "first", 1, 0).unwrap();
        assert_eq!(first.id, 1);
        let sec = s.insert_section_after(DIR, 4, "End").unwrap();
        assert_eq!(sec.id, 5);
        assert_eq!(texts(&s, DIR), ["first", "a", "b", "c", "End"]);
        assert!(s.insert_item_after(DIR, "x", 9, 0).is_err());
        assert!(s.insert_item_before(OTHER, "x", 1, 0).is_err());
    }

    #[test]
    fn active_items_list_first_in_position_order() {
        let (_tmp, s) = new_store();
        for t in ["a", "b", "c", "d"] {
            s.add(DIR, t).unwrap();
        }
        s.toggle_active(DIR, 4).unwrap();
        s.toggle_active(DIR, 2).unwrap();
        assert_eq!(texts(&s, DIR), ["b", "d", "a", "c"]);
        assert!(s.get(DIR, 2).unwrap().is_active);

        // Checking clears active; unchecking leaves it alone.
        s.check(DIR, 2).unwrap();
        assert!(!s.get(DIR, 2).unwrap().is_active);
        s.toggle(DIR, 4).unwrap();
        assert!(!s.get(DIR, 4).unwrap().is_active);
    }

    #[test]
    fn indent_is_clamped_between_zero_and_five() {
        let (_tmp, s) = new_store();
        s.add(DIR, "a").unwrap();
        for _ in 0..7 {
            s.indent(DIR, 1).unwrap();
        }
        assert_eq!(s.get(DIR, 1).unwrap().indent_level, 5);
        for _ in 0..7 {
            s.unindent(DIR, 1).unwrap();
        }
        assert_eq!(s.get(DIR, 1).unwrap().indent_level, 0);
    }

    #[test]
    fn markdown_round_trip() {
        let content = "# TODO\n\n## /home/me/project\n\n<!-- last-clean: 2026-01-02 -->\n- [ ] Fix bug\n  - [x] Reproduce @active\n    - [ ] Logs\n---\n--- Later\n- [ ] Deploy\n\n## /srv\n\n- [x] Done\n";
        let data = parse_markdown(content, Some("/home/me"));
        assert_eq!(data.directories[0].path, "/home/me/project");
        assert_eq!(data.directories[0].last_clean, "2026-01-02");
        let items = &data.directories[0].items;
        assert_eq!(items.len(), 6);
        assert!(items[1].checked && items[1].is_active);
        assert_eq!(items[1].text, "Reproduce");
        assert_eq!(items[2].indent_level, 2);
        assert!(items[3].is_section && items[3].text.is_empty());
        assert_eq!(items[4].text, "Later");

        let rendered = render_markdown(&data, Some("/home/me"));
        assert_eq!(
            rendered,
            content.replace("## /home/me/project", "## ~/project")
        );
    }

    #[test]
    fn parses_home_shorthand() {
        let data = parse_markdown("## ~\n- [ ] a\n## ~/x\n- [ ] b\n", Some("/h"));
        assert_eq!(data.directories[0].path, "/h");
        assert_eq!(data.directories[1].path, "/h/x");
        assert_eq!(
            render_markdown(&data, Some("/h")),
            "# TODO\n\n## ~\n\n- [ ] a\n\n## ~/x\n\n- [ ] b\n"
        );
    }

    #[test]
    fn clean_if_new_day() {
        let tmp = tempfile::tempdir().unwrap();
        let today = Rc::new(RefCell::new("2026-01-01".to_string()));
        let clock = today.clone();
        let s = Store::open(tmp.path().join("todo.md"))
            .unwrap()
            .with_today(move || clock.borrow().clone());

        // No directory yet: nothing recorded.
        s.clean_if_new_day(DIR).unwrap();
        assert!(!s.file_path.exists());

        s.add(DIR, "keep").unwrap();
        s.add(DIR, "done").unwrap();
        s.add(OTHER, "other done").unwrap();
        s.check(DIR, 2).unwrap();
        s.check(OTHER, 1).unwrap();

        // First run seeds the date without cleaning.
        s.clean_if_new_day(DIR).unwrap();
        assert_eq!(texts(&s, DIR), ["keep", "done"]);
        assert!(file(&s).contains("<!-- last-clean: 2026-01-01 -->"));

        // Same day: no change.
        s.clean_if_new_day(DIR).unwrap();
        assert_eq!(texts(&s, DIR), ["keep", "done"]);

        // New day: cleans only this directory and updates the date.
        *today.borrow_mut() = "2026-01-02".to_string();
        s.clean_if_new_day(DIR).unwrap();
        assert_eq!(texts(&s, DIR), ["keep"]);
        assert_eq!(texts(&s, OTHER), ["other done"]);
        assert!(file(&s).contains("<!-- last-clean: 2026-01-02 -->"));
    }

    #[test]
    fn clean_if_new_day_drops_emptied_directory() {
        let tmp = tempfile::tempdir().unwrap();
        let today = Rc::new(RefCell::new("2026-01-01".to_string()));
        let clock = today.clone();
        let s = Store::open(tmp.path().join("todo.md"))
            .unwrap()
            .with_today(move || clock.borrow().clone());
        s.add(DIR, "done").unwrap();
        s.check(DIR, 1).unwrap();
        s.clean_if_new_day(DIR).unwrap();
        *today.borrow_mut() = "2026-01-02".to_string();
        s.clean_if_new_day(DIR).unwrap();
        assert_eq!(file(&s), "# TODO\n\n");
    }
}
