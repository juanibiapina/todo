package tickets

import (
	"strings"
	"testing"
)

func TestFullStringMinimal(t *testing.T) {
	ticket := &Ticket{
		Title: "Fix login bug",
		ID:    "abc",
	}

	got := ticket.FullString()
	want := "---\nid: abc\n---\n# Fix login bug\n"

	if got != want {
		t.Errorf("FullString() =\n%q\nwant:\n%q", got, want)
	}
}

func TestFullStringWithDescription(t *testing.T) {
	ticket := &Ticket{
		Title:       "Fix login bug",
		ID:          "abc",
		Description: "The login page crashes on submit.",
	}

	got := ticket.FullString()
	want := "---\nid: abc\n---\n# Fix login bug\nThe login page crashes on submit.\n"

	if got != want {
		t.Errorf("FullString() =\n%q\nwant:\n%q", got, want)
	}
}

func TestFullStringWithAllFields(t *testing.T) {
	ticket := &Ticket{
		Title:       "Big feature",
		ID:          "XyZ",
		Description: "Implement the thing.",
		Status:      "open",
		Type:        "feature",
		Priority:    3,
		Assignee:    "alice",
		Created:     "2026-01-15",
		Parent:      "prt",
		ExternalRef: "JIRA-123",
		Deps:        []string{"dep1", "dep2"},
		Links:       []string{"lnk1"},
		Tags:        []string{"backend", "urgent"},
	}

	got := ticket.FullString()

	// Verify structure
	if !strings.HasPrefix(got, "---\n") {
		t.Error("should start with ---")
	}

	// Check all fields are present
	checks := []string{
		"id: XyZ",
		"status: open",
		"type: feature",
		"priority: 3",
		"assignee: alice",
		"created: \"2026-01-15\"",
		"parent: prt",
		"external_ref: JIRA-123",
		"deps:",
		"- dep1",
		"- dep2",
		"links:",
		"- lnk1",
		"tags:",
		"- backend",
		"- urgent",
		"# Big feature",
		"Implement the thing.",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("FullString() missing %q\ngot:\n%s", check, got)
		}
	}

	// Verify title comes after second ---
	parts := strings.SplitN(got, "---\n", 3)
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts split by ---, got %d", len(parts))
	}
	if !strings.HasPrefix(parts[2], "# Big feature\n") {
		t.Errorf("title should be first line after frontmatter, got:\n%s", parts[2])
	}
}

func TestFullStringWithArrays(t *testing.T) {
	ticket := &Ticket{
		Title: "Array test",
		ID:    "arr",
		Deps:  []string{"a", "b", "c"},
		Tags:  []string{"tag1"},
	}

	got := ticket.FullString()

	if !strings.Contains(got, "deps:\n    - a\n    - b\n    - c") {
		t.Errorf("FullString() deps format unexpected:\n%s", got)
	}
	if !strings.Contains(got, "tags:\n    - tag1") {
		t.Errorf("FullString() tags format unexpected:\n%s", got)
	}
}

func TestFullStringEmptyArrays(t *testing.T) {
	ticket := &Ticket{
		Title: "Empty arrays",
		ID:    "emp",
		Deps:  []string{},
		Links: []string{},
		Tags:  []string{},
	}

	got := ticket.FullString()

	// Empty slices should be omitted by omitempty
	if strings.Contains(got, "deps:") {
		t.Error("empty deps should be omitted")
	}
	if strings.Contains(got, "links:") {
		t.Error("empty links should be omitted")
	}
	if strings.Contains(got, "tags:") {
		t.Error("empty tags should be omitted")
	}
}

func TestFullStringNilArrays(t *testing.T) {
	ticket := &Ticket{
		Title: "Nil arrays",
		ID:    "nil",
	}

	got := ticket.FullString()

	if strings.Contains(got, "deps:") {
		t.Error("nil deps should be omitted")
	}
	if strings.Contains(got, "links:") {
		t.Error("nil links should be omitted")
	}
	if strings.Contains(got, "tags:") {
		t.Error("nil tags should be omitted")
	}
}

func TestFullStringPriorityZero(t *testing.T) {
	ticket := &Ticket{
		Title:    "Priority zero",
		ID:       "pz0",
		Priority: 0,
	}

	got := ticket.FullString()

	// Priority 0 is omitted by omitempty (acceptable per plan)
	if strings.Contains(got, "priority:") {
		t.Error("priority 0 should be omitted by omitempty")
	}
}

func TestFullStringPriorityNonZero(t *testing.T) {
	ticket := &Ticket{
		Title:    "Priority two",
		ID:       "pt2",
		Priority: 2,
	}

	got := ticket.FullString()

	if !strings.Contains(got, "priority: 2") {
		t.Errorf("should contain priority: 2, got:\n%s", got)
	}
}

func TestFullStringOmitsEmptyStrings(t *testing.T) {
	ticket := &Ticket{
		Title: "Minimal",
		ID:    "min",
	}

	got := ticket.FullString()

	omitted := []string{"status:", "type:", "assignee:", "created:", "parent:", "external_ref:"}
	for _, field := range omitted {
		if strings.Contains(got, field) {
			t.Errorf("empty field %q should be omitted, got:\n%s", field, got)
		}
	}
}

func TestFullStringMultilineDescription(t *testing.T) {
	ticket := &Ticket{
		Title:       "Multiline",
		ID:          "mlt",
		Description: "Line one.\n\nLine three.\n\n```go\nfmt.Println(\"hello\")\n```",
	}

	got := ticket.FullString()

	if !strings.Contains(got, "Line one.\n\nLine three.") {
		t.Error("multiline description not preserved")
	}
	if !strings.Contains(got, "```go\nfmt.Println(\"hello\")\n```") {
		t.Error("code block in description not preserved")
	}
}

func TestString(t *testing.T) {
	ticket := &Ticket{
		Title: "My ticket",
		ID:    "abc",
	}

	got := ticket.String()
	want := "abc My ticket"

	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestStringIgnoresOtherFields(t *testing.T) {
	ticket := &Ticket{
		Title:       "My ticket",
		ID:          "abc",
		Description: "Some desc",
		Status:      "open",
		Priority:    3,
	}

	got := ticket.String()
	want := "abc My ticket"

	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
