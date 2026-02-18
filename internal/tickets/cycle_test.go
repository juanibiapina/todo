package tickets

import (
	"testing"
)

func TestDepCyclesNoCycles(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	writeFile(dir, &Ticket{ID: "aaa", Title: "Root", Deps: []string{"bbb"}})
	writeFile(dir, &Ticket{ID: "bbb", Title: "Child"})

	result, err := DepCycles(dir)
	if err != nil {
		t.Fatalf("DepCycles: %v", err)
	}

	if result != "" {
		t.Errorf("expected empty string, got:\n%s", result)
	}
}

func TestDepCyclesSimple(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	writeFile(dir, &Ticket{ID: "aaa", Title: "First", Deps: []string{"bbb"}})
	writeFile(dir, &Ticket{ID: "bbb", Title: "Second", Deps: []string{"aaa"}})

	result, err := DepCycles(dir)
	if err != nil {
		t.Fatalf("DepCycles: %v", err)
	}

	expected := "Cycle: aaa -> bbb -> aaa\n  aaa First\n  bbb Second"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestDepCyclesThreeNode(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	writeFile(dir, &Ticket{ID: "aaa", Title: "A", Deps: []string{"bbb"}})
	writeFile(dir, &Ticket{ID: "bbb", Title: "B", Deps: []string{"ccc"}})
	writeFile(dir, &Ticket{ID: "ccc", Title: "C", Deps: []string{"aaa"}})

	result, err := DepCycles(dir)
	if err != nil {
		t.Fatalf("DepCycles: %v", err)
	}

	expected := "Cycle: aaa -> bbb -> ccc -> aaa\n  aaa A\n  bbb B\n  ccc C"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestDepCyclesMultiple(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	// Two independent cycles: aaa <-> bbb, ccc <-> ddd
	writeFile(dir, &Ticket{ID: "aaa", Title: "A", Deps: []string{"bbb"}})
	writeFile(dir, &Ticket{ID: "bbb", Title: "B", Deps: []string{"aaa"}})
	writeFile(dir, &Ticket{ID: "ccc", Title: "C", Deps: []string{"ddd"}})
	writeFile(dir, &Ticket{ID: "ddd", Title: "D", Deps: []string{"ccc"}})

	result, err := DepCycles(dir)
	if err != nil {
		t.Fatalf("DepCycles: %v", err)
	}

	expected := "Cycle: aaa -> bbb -> aaa\n  aaa A\n  bbb B\n\nCycle: ccc -> ddd -> ccc\n  ccc C\n  ddd D"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestDepCyclesClosedExcluded(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	// aaa -> bbb -> aaa, but bbb is closed — should break the cycle
	writeFile(dir, &Ticket{ID: "aaa", Title: "Open", Status: "open", Deps: []string{"bbb"}})
	writeFile(dir, &Ticket{ID: "bbb", Title: "Closed", Status: "closed", Deps: []string{"aaa"}})

	result, err := DepCycles(dir)
	if err != nil {
		t.Fatalf("DepCycles: %v", err)
	}

	if result != "" {
		t.Errorf("expected no cycles (closed ticket excluded), got:\n%s", result)
	}
}

func TestDepCyclesNormalized(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	// Cycle: ccc -> aaa -> bbb -> ccc
	// After normalization, should start with aaa
	writeFile(dir, &Ticket{ID: "ccc", Title: "C", Deps: []string{"aaa"}})
	writeFile(dir, &Ticket{ID: "aaa", Title: "A", Deps: []string{"bbb"}})
	writeFile(dir, &Ticket{ID: "bbb", Title: "B", Deps: []string{"ccc"}})

	result, err := DepCycles(dir)
	if err != nil {
		t.Fatalf("DepCycles: %v", err)
	}

	expected := "Cycle: aaa -> bbb -> ccc -> aaa\n  aaa A\n  bbb B\n  ccc C"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

func TestDepCyclesNoDeps(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	writeFile(dir, &Ticket{ID: "aaa", Title: "Solo"})
	writeFile(dir, &Ticket{ID: "bbb", Title: "Also solo"})

	result, err := DepCycles(dir)
	if err != nil {
		t.Fatalf("DepCycles: %v", err)
	}

	if result != "" {
		t.Errorf("expected empty string, got:\n%s", result)
	}
}

func TestDepCyclesWithStatus(t *testing.T) {
	dir := tempDir(t)
	EnsureDir(dir)

	writeFile(dir, &Ticket{ID: "aaa", Title: "First", Status: "open", Deps: []string{"bbb"}})
	writeFile(dir, &Ticket{ID: "bbb", Title: "Second", Status: "in_progress", Deps: []string{"aaa"}})

	result, err := DepCycles(dir)
	if err != nil {
		t.Fatalf("DepCycles: %v", err)
	}

	expected := "Cycle: aaa -> bbb -> aaa\n  aaa [open] First\n  bbb [in_progress] Second"
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}
