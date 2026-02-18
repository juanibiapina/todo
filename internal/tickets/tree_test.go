package tickets

import (
	"testing"
)

func TestDepTreeNoDeps(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	writeFile(dir, &Ticket{ID: "aaa", Title: "Root ticket", Status: "open"})

	result, err := DepTree(dir, "aaa", false)
	if err != nil {
		t.Fatalf("DepTree: %v", err)
	}

	expected := "aaa [open] Root ticket"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestDepTreeSimple(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	writeFile(dir, &Ticket{ID: "aaa", Title: "Root", Deps: []string{"bbb"}})
	writeFile(dir, &Ticket{ID: "bbb", Title: "Child"})

	result, err := DepTree(dir, "aaa", false)
	if err != nil {
		t.Fatalf("DepTree: %v", err)
	}

	expected := "aaa Root\n└── bbb Child"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestDepTreeNested(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	writeFile(dir, &Ticket{ID: "aaa", Title: "Root", Deps: []string{"bbb"}})
	writeFile(dir, &Ticket{ID: "bbb", Title: "Mid", Deps: []string{"ccc"}})
	writeFile(dir, &Ticket{ID: "ccc", Title: "Leaf"})

	result, err := DepTree(dir, "aaa", false)
	if err != nil {
		t.Fatalf("DepTree: %v", err)
	}

	expected := "aaa Root\n└── bbb Mid\n    └── ccc Leaf"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestDepTreeSorting(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	// aaa depends on bbb (depth 0) and ccc (depth 1, since ccc → ddd)
	writeFile(dir, &Ticket{ID: "aaa", Title: "Root", Deps: []string{"bbb", "ccc"}})
	writeFile(dir, &Ticket{ID: "bbb", Title: "Shallow"})
	writeFile(dir, &Ticket{ID: "ccc", Title: "Deep", Deps: []string{"ddd"}})
	writeFile(dir, &Ticket{ID: "ddd", Title: "Leaf"})

	result, err := DepTree(dir, "aaa", false)
	if err != nil {
		t.Fatalf("DepTree: %v", err)
	}

	// ccc (depth 1) should come before bbb (depth 0) because deeper first
	expected := "aaa Root\n├── ccc Deep\n│   └── ddd Leaf\n└── bbb Shallow"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestDepTreeCycle(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	writeFile(dir, &Ticket{ID: "aaa", Title: "Root", Deps: []string{"bbb"}})
	writeFile(dir, &Ticket{ID: "bbb", Title: "Child", Deps: []string{"aaa"}})

	result, err := DepTree(dir, "aaa", false)
	if err != nil {
		t.Fatalf("DepTree: %v", err)
	}

	expected := "aaa Root\n└── bbb Child\n    └── aaa Root (cycle)"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestDepTreeDedupDefault(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	// aaa → bbb, aaa → ccc, bbb → ccc (ccc appears in two branches)
	writeFile(dir, &Ticket{ID: "aaa", Title: "Root", Deps: []string{"bbb", "ccc"}})
	writeFile(dir, &Ticket{ID: "bbb", Title: "First", Deps: []string{"ccc"}})
	writeFile(dir, &Ticket{ID: "ccc", Title: "Shared"})

	result, err := DepTree(dir, "aaa", false)
	if err != nil {
		t.Fatalf("DepTree: %v", err)
	}

	// bbb has deeper subtree (depth 1) so comes first; ccc is expanded under bbb
	// The second ccc (direct dep of aaa) is a dup
	expected := "aaa Root\n├── bbb First\n│   └── ccc Shared\n└── ccc Shared (dup)"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestDepTreeFullNoDup(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	// Same structure as dedup test but with full=true
	writeFile(dir, &Ticket{ID: "aaa", Title: "Root", Deps: []string{"bbb", "ccc"}})
	writeFile(dir, &Ticket{ID: "bbb", Title: "First", Deps: []string{"ccc"}})
	writeFile(dir, &Ticket{ID: "ccc", Title: "Shared"})

	result, err := DepTree(dir, "aaa", true)
	if err != nil {
		t.Fatalf("DepTree: %v", err)
	}

	// In full mode, ccc appears fully in both places (no dup marker)
	expected := "aaa Root\n├── bbb First\n│   └── ccc Shared\n└── ccc Shared"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestDepTreeMissingDep(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	writeFile(dir, &Ticket{ID: "aaa", Title: "Root", Deps: []string{"bbb", "zzz"}})
	writeFile(dir, &Ticket{ID: "bbb", Title: "Valid"})
	// "zzz" doesn't exist — should be skipped

	result, err := DepTree(dir, "aaa", false)
	if err != nil {
		t.Fatalf("DepTree: %v", err)
	}

	expected := "aaa Root\n└── bbb Valid"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}
