//! Shared formatting helpers for displaying todo items across the CLI list
//! output and the TUI.

/// The three pieces of a section divider line, so callers can style each
/// piece independently (e.g. the title in a different colour from the dashes).
///
/// Concatenating `lead + mid + tail` equals [`section_line`]. If the title is
/// wider than the available width, the tail is a single space and the line
/// may exceed `width`.
pub fn section_parts(title: &str, width: usize) -> (String, String, String) {
    if title.is_empty() {
        return ("─".repeat(width), String::new(), String::new());
    }
    let lead = "── ".to_string();
    let prefix_width = lead.chars().count() + title.chars().count() + 1;
    let tail = if prefix_width >= width {
        " ".to_string()
    } else {
        format!(" {}", "─".repeat(width - prefix_width))
    };
    (lead, title.to_string(), tail)
}

/// Formats a section divider for display at the given width. With an empty
/// title it returns a row of box-drawing dashes. With a title it returns
/// `── <title> ` followed by dashes filling the rest of the width.
pub fn section_line(title: &str, width: usize) -> String {
    let (lead, mid, tail) = section_parts(title, width);
    lead + &mid + &tail
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn empty_title_is_all_dashes() {
        assert_eq!(section_line("", 10), "──────────");
        assert_eq!(section_line("", 0), "");
    }

    #[test]
    fn titled_section_fills_width() {
        assert_eq!(section_line("Plan", 20), "── Plan ────────────");
        assert_eq!(
            section_parts("Plan", 20),
            ("── ".into(), "Plan".into(), " ────────────".into())
        );
    }

    #[test]
    fn title_wider_than_width_omits_trailing_dashes() {
        assert_eq!(section_line("Long title", 5), "── Long title ");
    }
}
