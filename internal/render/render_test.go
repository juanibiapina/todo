package render

import "testing"

func TestSectionLineEmptyTitle(t *testing.T) {
	got := SectionLine("", 10)
	want := "──────────"
	if got != want {
		t.Errorf("SectionLine(\"\", 10) = %q, want %q", got, want)
	}
}

func TestSectionLineWithTitle(t *testing.T) {
	got := SectionLine("Plan", 20)
	want := "── Plan ────────────"
	if got != want {
		t.Errorf("SectionLine(\"Plan\", 20) = %q, want %q", got, want)
	}
}

func TestSectionLineTitleWiderThanWidth(t *testing.T) {
	// Title alone is wider than the requested width — no trailing dashes,
	// the prefix is returned as-is.
	got := SectionLine("Long title", 5)
	want := "── Long title "
	if got != want {
		t.Errorf("SectionLine(\"Long title\", 5) = %q, want %q", got, want)
	}
}

func TestSectionLineZeroWidth(t *testing.T) {
	if got := SectionLine("", 0); got != "" {
		t.Errorf("SectionLine(\"\", 0) = %q, want empty", got)
	}
}

func TestSectionPartsEmptyTitle(t *testing.T) {
	lead, mid, tail := SectionParts("", 10)
	if lead != "──────────" {
		t.Errorf("lead = %q, want 10 dashes", lead)
	}
	if mid != "" {
		t.Errorf("mid = %q, want empty", mid)
	}
	if tail != "" {
		t.Errorf("tail = %q, want empty", tail)
	}
}

func TestSectionPartsWithTitle(t *testing.T) {
	lead, mid, tail := SectionParts("Plan", 20)
	if lead != "── " {
		t.Errorf("lead = %q, want %q", lead, "── ")
	}
	if mid != "Plan" {
		t.Errorf("mid = %q, want %q", mid, "Plan")
	}
	if tail != " ────────────" {
		t.Errorf("tail = %q, want %q", tail, " ────────────")
	}
}

func TestSectionPartsConcatenateToSectionLine(t *testing.T) {
	cases := []struct {
		title string
		width int
	}{
		{"", 10},
		{"Plan", 20},
		{"Long title", 5},
		{"", 0},
	}
	for _, c := range cases {
		lead, mid, tail := SectionParts(c.title, c.width)
		got := lead + mid + tail
		want := SectionLine(c.title, c.width)
		if got != want {
			t.Errorf("SectionParts(%q, %d) concat = %q, SectionLine = %q", c.title, c.width, got, want)
		}
	}
}
