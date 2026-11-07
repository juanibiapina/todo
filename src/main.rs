mod config;
mod render;
mod store;
mod tui;

use std::error::Error;
use std::io;
use std::path::PathBuf;
use std::process::ExitCode;

use clap::{CommandFactory, Parser, Subcommand};
use clap_complete::Shell;

use render::section_line;
use store::Store;

const VERSION: &str = env!("CARGO_PKG_VERSION");

/// Width used for section dividers in CLI output.
const SECTION_WIDTH: usize = 30;

type CmdResult = Result<(), Box<dyn Error>>;

#[derive(Parser)]
#[command(
    name = "todo",
    about = "A simple per-directory todo list",
    long_about = "A CLI for managing simple todo items stored in a Markdown file, scoped per directory.",
    disable_version_flag = true
)]
struct Cli {
    /// Print version information
    #[arg(short = 'v', long = "version")]
    version: bool,

    #[command(subcommand)]
    command: Option<Command>,
}

#[derive(Subcommand)]
enum Command {
    /// Add a new todo item
    #[command(long_about = "Add a new unchecked todo item or section divider with the given text.")]
    Add {
        /// Add a section divider instead of a todo item
        #[arg(short, long)]
        section: bool,

        #[arg(value_name = "TEXT")]
        text: Vec<String>,
    },

    /// Check a todo item
    #[command(long_about = "Mark a todo item as checked.")]
    Check { id: String },

    /// Delete all checked items
    #[command(long_about = "Remove all checked todo items for the current directory.")]
    Clean,

    /// Edit a todo item
    #[command(long_about = "Replace the text of a todo item.")]
    Edit {
        id: String,

        #[arg(value_name = "TEXT", required = true)]
        text: Vec<String>,
    },

    /// List todo items
    #[command(long_about = "List all todo items for the current directory.")]
    List,

    /// Interactive todo list
    #[command(long_about = "Open an interactive terminal UI for managing todo items.")]
    Tui,

    /// Uncheck a todo item
    #[command(long_about = "Mark a todo item as unchecked.")]
    Uncheck { id: String },

    /// Generate the autocompletion script for the specified shell
    Completion { shell: Shell },
}

fn main() -> ExitCode {
    let cli = match Cli::try_parse() {
        Ok(cli) => cli,
        Err(e) => {
            let _ = e.print();
            return if e.use_stderr() {
                ExitCode::FAILURE
            } else {
                ExitCode::SUCCESS
            };
        }
    };

    if cli.version {
        println!("todo version {VERSION}");
        return ExitCode::SUCCESS;
    }

    match run(cli.command.unwrap_or(Command::List)) {
        Ok(()) => ExitCode::SUCCESS,
        Err(e) => {
            eprintln!("Error: {e}");
            ExitCode::FAILURE
        }
    }
}

fn run(command: Command) -> CmdResult {
    if let Command::Completion { shell } = command {
        clap_complete::generate(shell, &mut Cli::command(), "todo", &mut io::stdout());
        return Ok(());
    }

    let dir = working_dir()?;
    let store = Store::open(config::default_file_path())?;

    match command {
        Command::Add {
            section: true,
            text,
        } => {
            store.add_section(&dir, text.join(" ").trim())?;
        }
        Command::Add {
            section: false,
            text,
        } => {
            if text.is_empty() {
                return Err("requires at least 1 arg(s), only received 0".into());
            }
            let item = store.add(&dir, &text.join(" "))?;
            println!("{} {}", item.id, item.text);
        }
        Command::Check { id } => {
            let id = parse_id(&id)?;
            store.check(&dir, id)?;
            let item = store.get(&dir, id)?;
            println!("{} [x] {}", item.id, item.text);
        }
        Command::Uncheck { id } => {
            let id = parse_id(&id)?;
            store.uncheck(&dir, id)?;
            let item = store.get(&dir, id)?;
            println!("{} [ ] {}", item.id, item.text);
        }
        Command::Edit { id, text } => {
            let id = parse_id(&id)?;
            store.edit(&dir, id, &text.join(" "))?;
            let item = store.get(&dir, id)?;
            if item.is_section {
                println!("{} {}", item.id, section_line(&item.text, SECTION_WIDTH));
            } else {
                let check = if item.checked { "x" } else { " " };
                println!("{} [{check}] {}", item.id, item.text);
            }
        }
        Command::Clean => {
            let n = store.clean(&dir)?;
            println!("Deleted {n} item(s)");
        }
        Command::List => {
            store.clean_if_new_day(&dir)?;
            for item in store.list(&dir)? {
                if item.is_section {
                    println!("{}", section_line(&item.text, SECTION_WIDTH));
                    continue;
                }
                let check = if item.checked { "x" } else { " " };
                let active = if item.is_active { "  (active)" } else { "" };
                let indent = "  ".repeat(item.indent_level);
                println!("{} {indent}[{check}] {}{active}", item.id, item.text);
            }
        }
        Command::Tui => {
            store.clean_if_new_day(&dir)?;
            tui::run(store, dir)?;
        }
        Command::Completion { .. } => unreachable!("handled above"),
    }

    Ok(())
}

fn parse_id(arg: &str) -> Result<i64, String> {
    arg.parse().map_err(|_| format!("invalid item ID: {arg}"))
}

/// Returns the working directory, preferring `$PWD` when it names the same
/// directory so paths through symlinks keep the form the user typed.
fn working_dir() -> io::Result<String> {
    let cwd = std::env::current_dir()?;

    #[cfg(unix)]
    if let Some(pwd) = std::env::var_os("PWD").map(PathBuf::from)
        && pwd.is_absolute()
        && same_file(&pwd, &cwd)
    {
        return Ok(pwd.to_string_lossy().into_owned());
    }

    Ok(cwd.to_string_lossy().into_owned())
}

#[cfg(unix)]
fn same_file(a: &std::path::Path, b: &std::path::Path) -> bool {
    use std::os::unix::fs::MetadataExt;
    match (std::fs::metadata(a), std::fs::metadata(b)) {
        (Ok(a), Ok(b)) => a.dev() == b.dev() && a.ino() == b.ino(),
        _ => false,
    }
}
