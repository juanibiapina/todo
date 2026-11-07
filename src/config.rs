use std::env;
use std::path::PathBuf;

/// Returns the path to the todo Markdown file.
///
/// Resolution order:
///  1. `TODO_FILE` environment variable (supports `~` and `$VAR` expansion)
///  2. Default: `~/.local/share/todo/todo.md`
pub fn default_file_path() -> PathBuf {
    if let Ok(value) = env::var("TODO_FILE")
        && !value.is_empty()
    {
        return expand_path(&value);
    }

    match home_dir() {
        Some(home) => home
            .join(".local")
            .join("share")
            .join("todo")
            .join("todo.md"),
        None => PathBuf::from("todo.md"),
    }
}

/// Returns `$HOME` when it is set and non-empty.
pub fn home_dir() -> Option<PathBuf> {
    env::var_os("HOME")
        .filter(|h| !h.is_empty())
        .map(PathBuf::from)
}

/// Expands environment variables and a leading `~/` in a path.
fn expand_path(path: &str) -> PathBuf {
    let path = expand_env(path);
    if let Some(rest) = path.strip_prefix("~/")
        && let Some(home) = home_dir()
    {
        return home.join(rest);
    }
    PathBuf::from(path)
}

/// Replaces `$VAR` and `${VAR}` with the variable's value, or an empty string
/// when unset.
fn expand_env(input: &str) -> String {
    let mut out = String::with_capacity(input.len());
    let mut rest = input;

    while let Some(idx) = rest.find('$') {
        out.push_str(&rest[..idx]);
        let after = &rest[idx + 1..];
        match var_reference(after) {
            Some((name, len)) => {
                out.push_str(&env::var(name).unwrap_or_default());
                rest = &after[len..];
            }
            None => {
                out.push('$');
                rest = after;
            }
        }
    }

    out.push_str(rest);
    out
}

/// Parses the variable reference that follows a `$`: `{NAME}` or a run of
/// alphanumerics and underscores. Returns the name and the bytes consumed.
fn var_reference(s: &str) -> Option<(&str, usize)> {
    if let Some(braced) = s.strip_prefix('{') {
        let close = braced.find('}')?;
        return Some((&braced[..close], close + 2));
    }
    let len = s
        .find(|c: char| c != '_' && !c.is_ascii_alphanumeric())
        .unwrap_or(s.len());
    (len > 0).then(|| (&s[..len], len))
}

#[cfg(test)]
mod tests {
    use super::*;

    // Environment variables are process-global, so all env-dependent cases
    // run in one test to avoid races between parallel tests.
    #[test]
    fn resolves_todo_file() {
        let home = home_dir().expect("HOME must be set for tests");

        unsafe { env::set_var("TODO_FILE", "/custom/path/todo.md") };
        assert_eq!(default_file_path(), PathBuf::from("/custom/path/todo.md"));

        unsafe { env::set_var("TODO_TEST_MY_DIR", "/my/custom/dir") };
        unsafe { env::set_var("TODO_FILE", "$TODO_TEST_MY_DIR/todo.md") };
        assert_eq!(default_file_path(), PathBuf::from("/my/custom/dir/todo.md"));

        unsafe { env::set_var("TODO_FILE", "${TODO_TEST_MY_DIR}/x.md") };
        assert_eq!(default_file_path(), PathBuf::from("/my/custom/dir/x.md"));

        unsafe { env::set_var("TODO_FILE", "~/todo.md") };
        assert_eq!(default_file_path(), home.join("todo.md"));

        unsafe { env::set_var("TODO_FILE", "") };
        assert_eq!(default_file_path(), home.join(".local/share/todo/todo.md"));

        unsafe { env::remove_var("TODO_FILE") };
        unsafe { env::remove_var("TODO_TEST_MY_DIR") };
    }
}
