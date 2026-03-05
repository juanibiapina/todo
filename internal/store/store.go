package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Item represents a single todo item.
type Item struct {
	ID        int
	Directory string
	Text      string
	Checked   bool
}

// Store provides access to the todo database.
type Store struct {
	db *sql.DB
}

// DefaultDBPath returns the default database path, respecting XDG_DATA_HOME.
func DefaultDBPath() string {
	if v := os.Getenv("TODO_DB"); v != "" {
		return v
	}
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, _ := os.UserHomeDir()
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "todo", "todo.db")
}

// Open opens (or creates) the database at the given path.
func Open(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrating database: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			directory TEXT NOT NULL,
			text TEXT NOT NULL,
			checked INTEGER NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_items_directory ON items(directory);
	`)
	return err
}

// Add inserts a new unchecked item for the given directory.
func (s *Store) Add(directory, text string) (*Item, error) {
	result, err := s.db.Exec(
		"INSERT INTO items (directory, text, checked) VALUES (?, ?, 0)",
		directory, text,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Item{
		ID:        int(id),
		Directory: directory,
		Text:      text,
		Checked:   false,
	}, nil
}

// List returns all items for the given directory, ordered by ID.
func (s *Store) List(directory string) ([]*Item, error) {
	rows, err := s.db.Query(
		"SELECT id, directory, text, checked FROM items WHERE directory = ? ORDER BY id",
		directory,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var item Item
		var checked int
		if err := rows.Scan(&item.ID, &item.Directory, &item.Text, &checked); err != nil {
			return nil, err
		}
		item.Checked = checked != 0
		items = append(items, &item)
	}
	return items, rows.Err()
}

// Check marks an item as checked. Returns an error if the item doesn't exist
// or doesn't belong to the given directory.
func (s *Store) Check(directory string, id int) error {
	return s.setChecked(directory, id, true)
}

// Uncheck marks an item as unchecked. Returns an error if the item doesn't exist
// or doesn't belong to the given directory.
func (s *Store) Uncheck(directory string, id int) error {
	return s.setChecked(directory, id, false)
}

// Toggle flips the checked state of an item. Returns the new state.
func (s *Store) Toggle(directory string, id int) (bool, error) {
	item, err := s.Get(directory, id)
	if err != nil {
		return false, err
	}

	newState := !item.Checked
	if err := s.setChecked(directory, id, newState); err != nil {
		return false, err
	}
	return newState, nil
}

// Get returns a single item by ID and directory.
func (s *Store) Get(directory string, id int) (*Item, error) {
	var item Item
	var checked int
	err := s.db.QueryRow(
		"SELECT id, directory, text, checked FROM items WHERE id = ? AND directory = ?",
		id, directory,
	).Scan(&item.ID, &item.Directory, &item.Text, &checked)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("item %d not found", id)
	}
	if err != nil {
		return nil, err
	}
	item.Checked = checked != 0
	return &item, nil
}

func (s *Store) setChecked(directory string, id int, checked bool) error {
	val := 0
	if checked {
		val = 1
	}
	result, err := s.db.Exec(
		"UPDATE items SET checked = ? WHERE id = ? AND directory = ?",
		val, id, directory,
	)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("item %d not found", id)
	}
	return nil
}

// Edit updates the text of an item. Returns an error if the item doesn't exist
// or doesn't belong to the given directory.
func (s *Store) Edit(directory string, id int, text string) error {
	result, err := s.db.Exec(
		"UPDATE items SET text = ? WHERE id = ? AND directory = ?",
		text, id, directory,
	)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("item %d not found", id)
	}
	return nil
}

// Clean deletes all checked items for the given directory. Returns the number deleted.
func (s *Store) Clean(directory string) (int, error) {
	result, err := s.db.Exec(
		"DELETE FROM items WHERE directory = ? AND checked = 1",
		directory,
	)
	if err != nil {
		return 0, err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(n), nil
}
