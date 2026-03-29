package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Item represents a single todo item or section divider.
type Item struct {
	ID          int
	Directory   string
	Text        string
	Checked     bool
	IsSection   bool
	IsActive    bool
	IndentLevel int
}

// Store provides access to the todo file.
type Store struct {
	filePath string
	now      func() time.Time
}

// Internal data model

type fileData struct {
	directories []*directoryData
}

type directoryData struct {
	path      string // expanded (absolute) path
	lastClean string // "YYYY-MM-DD" or "" if unset
	items     []*fileItem
}

type fileItem struct {
	text        string
	checked     bool
	isSection   bool
	isActive    bool
	indentLevel int
}

// Open creates a Store backed by the given Markdown file path.
// Creates the parent directory if needed.
func Open(filePath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, fmt.Errorf("creating directory: %w", err)
	}
	return &Store{filePath: filePath, now: time.Now}, nil
}

// Close is a no-op for the Markdown backend.
func (s *Store) Close() error {
	return nil
}

// Add inserts a new unchecked item for the given directory.
func (s *Store) Add(directory, text string) (*Item, error) {
	data, err := s.readFile()
	if err != nil {
		return nil, err
	}

	dir := getOrCreateDirectory(data, directory)
	dir.items = append(dir.items, &fileItem{text: text})

	if err := s.writeFile(data); err != nil {
		return nil, err
	}

	id := len(dir.items)
	return &Item{
		ID:        id,
		Directory: directory,
		Text:      text,
	}, nil
}

// Indent increments the indent level of an item by one (max 3).
// Returns an error if the item is a section or doesn't exist.
func (s *Store) Indent(directory string, id int) error {
	return s.setIndentLevel(directory, id, 1)
}

// Unindent decrements the indent level of an item by one (min 0).
// Returns an error if the item is a section or doesn't exist.
func (s *Store) Unindent(directory string, id int) error {
	return s.setIndentLevel(directory, id, -1)
}

func (s *Store) setIndentLevel(directory string, id int, delta int) error {
	data, err := s.readFile()
	if err != nil {
		return err
	}

	dir := findDirectory(data, directory)
	if dir == nil || id < 1 || id > len(dir.items) {
		return fmt.Errorf("item %d not found", id)
	}

	fi := dir.items[id-1]
	if fi.isSection {
		return fmt.Errorf("item %d is a section", id)
	}

	newLevel := fi.indentLevel + delta
	if newLevel < 0 {
		newLevel = 0
	}
	if newLevel > 3 {
		newLevel = 3
	}
	fi.indentLevel = newLevel

	return s.writeFile(data)
}

// AddSection inserts a new section divider for the given directory.
func (s *Store) AddSection(directory string) (*Item, error) {
	data, err := s.readFile()
	if err != nil {
		return nil, err
	}

	dir := getOrCreateDirectory(data, directory)
	dir.items = append(dir.items, &fileItem{isSection: true})

	if err := s.writeFile(data); err != nil {
		return nil, err
	}

	id := len(dir.items)
	return &Item{
		ID:        id,
		Directory: directory,
		IsSection: true,
	}, nil
}

// InsertItemAfter inserts a new item right after the given position.
func (s *Store) InsertItemAfter(directory, text string, afterID int) (*Item, error) {
	return s.insertAfter(directory, text, afterID, false)
}

// InsertSectionAfter inserts a new section divider right after the given position.
func (s *Store) InsertSectionAfter(directory string, afterID int) (*Item, error) {
	return s.insertAfter(directory, "", afterID, true)
}

func (s *Store) insertAfter(directory, text string, afterID int, isSection bool) (*Item, error) {
	data, err := s.readFile()
	if err != nil {
		return nil, err
	}

	dir := findDirectory(data, directory)
	if dir == nil || afterID < 1 || afterID > len(dir.items) {
		return nil, fmt.Errorf("item %d not found", afterID)
	}

	newItem := &fileItem{text: text, isSection: isSection}

	// Insert after afterID position (afterID is 1-based)
	items := make([]*fileItem, 0, len(dir.items)+1)
	items = append(items, dir.items[:afterID]...)
	items = append(items, newItem)
	items = append(items, dir.items[afterID:]...)
	dir.items = items

	if err := s.writeFile(data); err != nil {
		return nil, err
	}

	newID := afterID + 1
	return &Item{
		ID:          newID,
		Directory:   directory,
		Text:        text,
		IsSection:   isSection,
	}, nil
}

// List returns all items for the given directory, ordered by active status
// (active first) then position.
func (s *Store) List(directory string) ([]*Item, error) {
	data, err := s.readFile()
	if err != nil {
		return nil, err
	}

	dir := findDirectory(data, directory)
	if dir == nil {
		return nil, nil
	}

	items := make([]*Item, len(dir.items))
	for i, fi := range dir.items {
		items[i] = &Item{
			ID:          i + 1,
			Directory:   directory,
			Text:        fi.text,
			Checked:     fi.checked,
			IsSection:   fi.isSection,
			IsActive:    fi.isActive,
			IndentLevel: fi.indentLevel,
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].IsActive != items[j].IsActive {
			return items[i].IsActive
		}
		return items[i].ID < items[j].ID
	})

	return items, nil
}

// Get returns a single item by position (1-based) within the directory.
func (s *Store) Get(directory string, id int) (*Item, error) {
	data, err := s.readFile()
	if err != nil {
		return nil, err
	}

	dir := findDirectory(data, directory)
	if dir == nil || id < 1 || id > len(dir.items) {
		return nil, fmt.Errorf("item %d not found", id)
	}

	fi := dir.items[id-1]
	return &Item{
		ID:          id,
		Directory:   directory,
		Text:        fi.text,
		Checked:     fi.checked,
		IsSection:   fi.isSection,
		IsActive:    fi.isActive,
		IndentLevel: fi.indentLevel,
	}, nil
}

// Check marks an item as checked. Returns an error if the item doesn't exist
// or is a section.
func (s *Store) Check(directory string, id int) error {
	return s.setChecked(directory, id, true)
}

// Uncheck marks an item as unchecked. Returns an error if the item doesn't
// exist or is a section.
func (s *Store) Uncheck(directory string, id int) error {
	return s.setChecked(directory, id, false)
}

func (s *Store) setChecked(directory string, id int, checked bool) error {
	data, err := s.readFile()
	if err != nil {
		return err
	}

	dir := findDirectory(data, directory)
	if dir == nil || id < 1 || id > len(dir.items) {
		return fmt.Errorf("item %d not found", id)
	}

	fi := dir.items[id-1]
	if fi.isSection {
		return fmt.Errorf("item %d is a section", id)
	}

	fi.checked = checked
	if checked {
		fi.isActive = false
	}

	return s.writeFile(data)
}

// Toggle flips the checked state of an item. Returns the new state.
// Returns an error if the item is a section.
func (s *Store) Toggle(directory string, id int) (bool, error) {
	data, err := s.readFile()
	if err != nil {
		return false, err
	}

	dir := findDirectory(data, directory)
	if dir == nil || id < 1 || id > len(dir.items) {
		return false, fmt.Errorf("item %d not found", id)
	}

	fi := dir.items[id-1]
	if fi.isSection {
		return false, fmt.Errorf("item %d is a section", id)
	}

	fi.checked = !fi.checked
	if fi.checked {
		fi.isActive = false
	}

	if err := s.writeFile(data); err != nil {
		return false, err
	}
	return fi.checked, nil
}

// ToggleActive flips the active state of an item. Returns the new state.
// Returns an error if the item is a section.
func (s *Store) ToggleActive(directory string, id int) (bool, error) {
	data, err := s.readFile()
	if err != nil {
		return false, err
	}

	dir := findDirectory(data, directory)
	if dir == nil || id < 1 || id > len(dir.items) {
		return false, fmt.Errorf("item %d not found", id)
	}

	fi := dir.items[id-1]
	if fi.isSection {
		return false, fmt.Errorf("item %d is a section", id)
	}

	fi.isActive = !fi.isActive
	if err := s.writeFile(data); err != nil {
		return false, err
	}
	return fi.isActive, nil
}

// Edit updates the text of an item. Returns an error if the item doesn't exist
// or is a section.
func (s *Store) Edit(directory string, id int, text string) error {
	data, err := s.readFile()
	if err != nil {
		return err
	}

	dir := findDirectory(data, directory)
	if dir == nil || id < 1 || id > len(dir.items) {
		return fmt.Errorf("item %d not found", id)
	}

	fi := dir.items[id-1]
	if fi.isSection {
		return fmt.Errorf("item %d is a section", id)
	}

	fi.text = text
	return s.writeFile(data)
}

// Delete removes a single item by position. Returns an error if the item
// doesn't exist.
func (s *Store) Delete(directory string, id int) error {
	data, err := s.readFile()
	if err != nil {
		return err
	}

	dir := findDirectory(data, directory)
	if dir == nil || id < 1 || id > len(dir.items) {
		return fmt.Errorf("item %d not found", id)
	}

	dir.items = append(dir.items[:id-1], dir.items[id:]...)
	return s.writeFile(data)
}

// Swap exchanges two items at the given positions within the same directory.
func (s *Store) Swap(directory string, id1, id2 int) error {
	data, err := s.readFile()
	if err != nil {
		return err
	}

	dir := findDirectory(data, directory)
	if dir == nil {
		return fmt.Errorf("item %d not found", id1)
	}
	if id1 < 1 || id1 > len(dir.items) {
		return fmt.Errorf("item %d not found", id1)
	}
	if id2 < 1 || id2 > len(dir.items) {
		return fmt.Errorf("item %d not found", id2)
	}

	dir.items[id1-1], dir.items[id2-1] = dir.items[id2-1], dir.items[id1-1]
	return s.writeFile(data)
}

// Clean deletes all checked items for the given directory. Returns the count.
func (s *Store) Clean(directory string) (int, error) {
	data, err := s.readFile()
	if err != nil {
		return 0, err
	}

	dir := findDirectory(data, directory)
	if dir == nil {
		return 0, nil
	}

	var remaining []*fileItem
	count := 0
	for _, fi := range dir.items {
		if fi.checked && !fi.isSection {
			count++
		} else {
			remaining = append(remaining, fi)
		}
	}
	dir.items = remaining

	if err := s.writeFile(data); err != nil {
		return 0, err
	}
	return count, nil
}

// CleanIfNewDay automatically cleans checked items when a new day begins.
// On first use for a directory, it records today's date without cleaning.
// On subsequent days, it deletes all checked items before returning.
func (s *Store) CleanIfNewDay(directory string) error {
	data, err := s.readFile()
	if err != nil {
		return err
	}

	today := s.now().Format("2006-01-02")

	dir := findDirectory(data, directory)
	if dir == nil {
		// No items for this directory yet, nothing to track or clean
		return nil
	}

	if dir.lastClean == "" {
		// First use for existing directory: seed the date, don't clean
		dir.lastClean = today
		return s.writeFile(data)
	}

	if dir.lastClean == today {
		return nil
	}

	// New day: clean checked items and update date
	var remaining []*fileItem
	for _, fi := range dir.items {
		if fi.checked && !fi.isSection {
			continue
		}
		remaining = append(remaining, fi)
	}
	dir.items = remaining
	dir.lastClean = today
	return s.writeFile(data)
}

// File I/O

func (s *Store) readFile() (*fileData, error) {
	content, err := os.ReadFile(s.filePath)
	if os.IsNotExist(err) {
		return &fileData{}, nil
	}
	if err != nil {
		return nil, err
	}
	return parseMarkdown(string(content)), nil
}

func (s *Store) writeFile(data *fileData) error {
	removeEmptyDirectories(data)
	content := renderMarkdown(data)

	dir := filepath.Dir(s.filePath)
	tmp, err := os.CreateTemp(dir, ".todo-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	_, writeErr := tmp.WriteString(content)
	closeErr := tmp.Close()

	if writeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("writing temp file: %w", writeErr)
	}
	if closeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", closeErr)
	}

	if err := os.Rename(tmpPath, s.filePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}

// Markdown parsing and rendering

func parseMarkdown(content string) *fileData {
	data := &fileData{}
	var current *directoryData

	home, _ := os.UserHomeDir()

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)

		// Top-level heading (preserved, not parsed as data)
		if strings.HasPrefix(trimmed, "# ") && !strings.HasPrefix(trimmed, "## ") {
			continue
		}

		// Directory heading
		if strings.HasPrefix(trimmed, "## ") {
			path := strings.TrimPrefix(trimmed, "## ")
			if path == "~" {
				path = home
			} else if strings.HasPrefix(path, "~/") {
				path = filepath.Join(home, path[2:])
			}
			current = &directoryData{path: path}
			data.directories = append(data.directories, current)
			continue
		}

		if current == nil {
			continue
		}

		// Last clean date
		if strings.HasPrefix(trimmed, "<!-- last-clean: ") && strings.HasSuffix(trimmed, " -->") {
			date := strings.TrimPrefix(trimmed, "<!-- last-clean: ")
			date = strings.TrimSuffix(date, " -->")
			current.lastClean = date
			continue
		}

		// Section divider
		if trimmed == "---" {
			current.items = append(current.items, &fileItem{isSection: true})
			continue
		}

		// Unchecked item
		if strings.HasPrefix(trimmed, "- [ ] ") {
			indentLevel := countLeadingSpaces(line) / 2
			text := strings.TrimPrefix(trimmed, "- [ ] ")
			isActive := false
			if strings.HasSuffix(text, " @active") {
				text = strings.TrimSuffix(text, " @active")
				isActive = true
			}
			current.items = append(current.items, &fileItem{
				text:        text,
				isActive:    isActive,
				indentLevel: indentLevel,
			})
			continue
		}

		// Checked item
		if strings.HasPrefix(trimmed, "- [x] ") {
			indentLevel := countLeadingSpaces(line) / 2
			text := strings.TrimPrefix(trimmed, "- [x] ")
			isActive := false
			if strings.HasSuffix(text, " @active") {
				text = strings.TrimSuffix(text, " @active")
				isActive = true
			}
			current.items = append(current.items, &fileItem{
				text:        text,
				checked:     true,
				isActive:    isActive,
				indentLevel: indentLevel,
			})
			continue
		}
	}

	return data
}

func renderMarkdown(data *fileData) string {
	home, _ := os.UserHomeDir()

	var b strings.Builder
	b.WriteString("# TODO\n\n")

	for i, dir := range data.directories {
		if i > 0 {
			b.WriteString("\n")
		}

		// Collapse home prefix to ~
		path := dir.path
		if path == home {
			path = "~"
		} else if strings.HasPrefix(path, home+"/") {
			path = "~/" + strings.TrimPrefix(path, home+"/")
		}

		b.WriteString("## ")
		b.WriteString(path)
		b.WriteString("\n\n")

		if dir.lastClean != "" {
			b.WriteString("<!-- last-clean: ")
			b.WriteString(dir.lastClean)
			b.WriteString(" -->\n")
		}

		for _, item := range dir.items {
			if item.isSection {
				b.WriteString("---\n")
			} else {
				check := "[ ]"
				if item.checked {
					check = "[x]"
				}
				b.WriteString(strings.Repeat("  ", item.indentLevel))
				b.WriteString("- ")
				b.WriteString(check)
				b.WriteString(" ")
				b.WriteString(item.text)
				if item.isActive {
					b.WriteString(" @active")
				}
				b.WriteString("\n")
			}
		}
	}

	return b.String()
}

// Helpers

func findDirectory(data *fileData, path string) *directoryData {
	for _, dir := range data.directories {
		if dir.path == path {
			return dir
		}
	}
	return nil
}

func removeEmptyDirectories(data *fileData) {
	var kept []*directoryData
	for _, dir := range data.directories {
		if len(dir.items) > 0 {
			kept = append(kept, dir)
		}
	}
	data.directories = kept
}

func getOrCreateDirectory(data *fileData, path string) *directoryData {
	if dir := findDirectory(data, path); dir != nil {
		return dir
	}
	dir := &directoryData{path: path}
	data.directories = append(data.directories, dir)
	return dir
}

func countLeadingSpaces(s string) int {
	for i, ch := range s {
		if ch != ' ' {
			return i
		}
	}
	return len(s)
}
