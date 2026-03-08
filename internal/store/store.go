package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
	db  *sql.DB
	now func() time.Time
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

	return &Store{db: db, now: time.Now}, nil
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
			checked INTEGER NOT NULL DEFAULT 0,
			position INTEGER NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_items_directory ON items(directory);
		CREATE TABLE IF NOT EXISTS daily_clean (
			directory TEXT PRIMARY KEY,
			last_date TEXT NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	// Add position column for existing databases.
	if !hasColumn(db, "items", "position") {
		_, err = db.Exec(`ALTER TABLE items ADD COLUMN position INTEGER NOT NULL DEFAULT 0`)
		if err != nil {
			return err
		}
	}

	// Initialize positions for any items that don't have one.
	_, err = db.Exec(`UPDATE items SET position = id WHERE position = 0`)
	return err
}

func hasColumn(db *sql.DB, table, column string) bool {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull int
		var dfltValue *string
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			return false
		}
		if name == column {
			return true
		}
	}
	return false
}

// Add inserts a new unchecked item for the given directory.
func (s *Store) Add(directory, text string) (*Item, error) {
	var maxPos sql.NullInt64
	s.db.QueryRow("SELECT MAX(position) FROM items WHERE directory = ?", directory).Scan(&maxPos)
	newPos := 1
	if maxPos.Valid {
		newPos = int(maxPos.Int64) + 1
	}

	result, err := s.db.Exec(
		"INSERT INTO items (directory, text, checked, position) VALUES (?, ?, 0, ?)",
		directory, text, newPos,
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
		"SELECT id, directory, text, checked FROM items WHERE directory = ? ORDER BY position",
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

// Delete removes a single item by ID and directory. Returns an error if the item
// doesn't exist or doesn't belong to the given directory.
func (s *Store) Delete(directory string, id int) error {
	result, err := s.db.Exec(
		"DELETE FROM items WHERE id = ? AND directory = ?",
		id, directory,
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

// Swap exchanges the positions of two items within the same directory.
func (s *Store) Swap(directory string, id1, id2 int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var pos1, pos2 int
	err = tx.QueryRow("SELECT position FROM items WHERE id = ? AND directory = ?", id1, directory).Scan(&pos1)
	if err != nil {
		return fmt.Errorf("item %d not found", id1)
	}
	err = tx.QueryRow("SELECT position FROM items WHERE id = ? AND directory = ?", id2, directory).Scan(&pos2)
	if err != nil {
		return fmt.Errorf("item %d not found", id2)
	}

	_, err = tx.Exec("UPDATE items SET position = ? WHERE id = ? AND directory = ?", pos2, id1, directory)
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE items SET position = ? WHERE id = ? AND directory = ?", pos1, id2, directory)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// CleanIfNewDay automatically cleans checked items when a new day begins.
// On first use for a directory, it records today's date without cleaning.
// On subsequent days, it deletes all checked items before returning.
func (s *Store) CleanIfNewDay(directory string) error {
	today := s.now().Format("2006-01-02")

	var lastDate string
	err := s.db.QueryRow(
		"SELECT last_date FROM daily_clean WHERE directory = ?",
		directory,
	).Scan(&lastDate)

	if err == sql.ErrNoRows {
		// First use: record today, don't clean
		_, err = s.db.Exec(
			"INSERT INTO daily_clean (directory, last_date) VALUES (?, ?)",
			directory, today,
		)
		return err
	}
	if err != nil {
		return err
	}

	if lastDate == today {
		return nil
	}

	// New day: clean checked items and update date
	if _, err := s.Clean(directory); err != nil {
		return err
	}
	_, err = s.db.Exec(
		"UPDATE daily_clean SET last_date = ? WHERE directory = ?",
		today, directory,
	)
	return err
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
