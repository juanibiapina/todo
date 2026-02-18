package tickets

import (
	"strings"
	"testing"
)

func TestComputeRelations_NoRelations(t *testing.T) {
	ticket := &Ticket{ID: "aaa", Title: "Standalone"}
	all := []*Ticket{ticket}

	rel := ComputeRelations(ticket, all)

	if rel.ParentTicket != nil {
		t.Errorf("expected nil parent, got %v", rel.ParentTicket)
	}
	if len(rel.Blockers) != 0 {
		t.Errorf("expected 0 blockers, got %d", len(rel.Blockers))
	}
	if len(rel.Blocking) != 0 {
		t.Errorf("expected 0 blocking, got %d", len(rel.Blocking))
	}
	if len(rel.Children) != 0 {
		t.Errorf("expected 0 children, got %d", len(rel.Children))
	}
	if len(rel.Linked) != 0 {
		t.Errorf("expected 0 linked, got %d", len(rel.Linked))
	}
}

func TestComputeRelations_Parent(t *testing.T) {
	parent := &Ticket{ID: "par", Title: "Parent Ticket"}
	child := &Ticket{ID: "chi", Title: "Child Ticket", Parent: "par"}
	all := []*Ticket{parent, child}

	rel := ComputeRelations(child, all)

	if rel.ParentTicket == nil {
		t.Fatal("expected parent ticket, got nil")
	}
	if rel.ParentTicket.ID != "par" {
		t.Errorf("expected parent ID 'par', got %q", rel.ParentTicket.ID)
	}
	if rel.ParentTicket.Title != "Parent Ticket" {
		t.Errorf("expected parent title 'Parent Ticket', got %q", rel.ParentTicket.Title)
	}
}

func TestComputeRelations_ParentNotFound(t *testing.T) {
	ticket := &Ticket{ID: "aaa", Title: "Orphan", Parent: "zzz"}
	all := []*Ticket{ticket}

	rel := ComputeRelations(ticket, all)

	if rel.ParentTicket != nil {
		t.Errorf("expected nil parent for missing parent ID, got %v", rel.ParentTicket)
	}
}

func TestComputeRelations_Blockers(t *testing.T) {
	dep1 := &Ticket{ID: "d1a", Title: "Open Dep", Status: "open"}
	dep2 := &Ticket{ID: "d2a", Title: "Closed Dep", Status: "closed"}
	dep3 := &Ticket{ID: "d3a", Title: "In Progress Dep", Status: "in_progress"}
	ticket := &Ticket{ID: "aaa", Title: "Main", Deps: []string{"d1a", "d2a", "d3a"}}
	all := []*Ticket{dep1, dep2, dep3, ticket}

	rel := ComputeRelations(ticket, all)

	if len(rel.Blockers) != 2 {
		t.Fatalf("expected 2 blockers (open + in_progress), got %d", len(rel.Blockers))
	}
	if rel.Blockers[0].ID != "d1a" {
		t.Errorf("expected first blocker 'd1a', got %q", rel.Blockers[0].ID)
	}
	if rel.Blockers[1].ID != "d3a" {
		t.Errorf("expected second blocker 'd3a', got %q", rel.Blockers[1].ID)
	}
}

func TestComputeRelations_BlockersMissingDep(t *testing.T) {
	ticket := &Ticket{ID: "aaa", Title: "Main", Deps: []string{"zzz"}}
	all := []*Ticket{ticket}

	rel := ComputeRelations(ticket, all)

	// Missing deps are not listed as blockers
	if len(rel.Blockers) != 0 {
		t.Errorf("expected 0 blockers for missing dep, got %d", len(rel.Blockers))
	}
}

func TestComputeRelations_Blocking(t *testing.T) {
	ticket := &Ticket{ID: "aaa", Title: "Blocker", Status: "open"}
	blocked1 := &Ticket{ID: "b1a", Title: "Blocked 1", Deps: []string{"aaa"}}
	blocked2 := &Ticket{ID: "b2a", Title: "Blocked 2", Deps: []string{"aaa"}}
	unrelated := &Ticket{ID: "uuu", Title: "Unrelated"}
	all := []*Ticket{ticket, blocked1, blocked2, unrelated}

	rel := ComputeRelations(ticket, all)

	if len(rel.Blocking) != 2 {
		t.Fatalf("expected 2 blocking, got %d", len(rel.Blocking))
	}
}

func TestComputeRelations_BlockingNotShownWhenClosed(t *testing.T) {
	ticket := &Ticket{ID: "aaa", Title: "Done", Status: "closed"}
	blocked := &Ticket{ID: "bbb", Title: "Depends on Done", Deps: []string{"aaa"}}
	all := []*Ticket{ticket, blocked}

	rel := ComputeRelations(ticket, all)

	// When the ticket is closed, it doesn't block anyone
	if len(rel.Blocking) != 0 {
		t.Errorf("expected 0 blocking when ticket is closed, got %d", len(rel.Blocking))
	}
}

func TestComputeRelations_Children(t *testing.T) {
	parent := &Ticket{ID: "par", Title: "Epic"}
	child1 := &Ticket{ID: "c1a", Title: "Task 1", Parent: "par"}
	child2 := &Ticket{ID: "c2a", Title: "Task 2", Parent: "par"}
	other := &Ticket{ID: "ooo", Title: "Other", Parent: "xxx"}
	all := []*Ticket{parent, child1, child2, other}

	rel := ComputeRelations(parent, all)

	if len(rel.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(rel.Children))
	}
	if rel.Children[0].ID != "c1a" {
		t.Errorf("expected first child 'c1a', got %q", rel.Children[0].ID)
	}
	if rel.Children[1].ID != "c2a" {
		t.Errorf("expected second child 'c2a', got %q", rel.Children[1].ID)
	}
}

func TestComputeRelations_Linked(t *testing.T) {
	ticket := &Ticket{ID: "aaa", Title: "Main", Links: []string{"bbb", "ccc"}}
	link1 := &Ticket{ID: "bbb", Title: "Linked 1"}
	link2 := &Ticket{ID: "ccc", Title: "Linked 2"}
	all := []*Ticket{ticket, link1, link2}

	rel := ComputeRelations(ticket, all)

	if len(rel.Linked) != 2 {
		t.Fatalf("expected 2 linked, got %d", len(rel.Linked))
	}
	if rel.Linked[0].ID != "bbb" {
		t.Errorf("expected first linked 'bbb', got %q", rel.Linked[0].ID)
	}
}

func TestComputeRelations_LinkedMissing(t *testing.T) {
	ticket := &Ticket{ID: "aaa", Title: "Main", Links: []string{"zzz"}}
	all := []*Ticket{ticket}

	rel := ComputeRelations(ticket, all)

	// Missing links are not listed
	if len(rel.Linked) != 0 {
		t.Errorf("expected 0 linked for missing link, got %d", len(rel.Linked))
	}
}

func TestFormatRelations_Empty(t *testing.T) {
	rel := &TicketRelations{}

	result := FormatRelations(rel)

	if result != "" {
		t.Errorf("expected empty string for no relations, got %q", result)
	}
}

func TestFormatRelations_AllSections(t *testing.T) {
	rel := &TicketRelations{
		Blockers: []*Ticket{
			{ID: "b1a", Title: "Blocker One", Status: "open"},
		},
		Blocking: []*Ticket{
			{ID: "b2a", Title: "Blocked By Me", Status: "in_progress"},
		},
		Children: []*Ticket{
			{ID: "c1a", Title: "Child Task"},
		},
		Linked: []*Ticket{
			{ID: "l1a", Title: "Related", Status: "open"},
		},
	}

	result := FormatRelations(rel)

	if !strings.Contains(result, "## Blockers") {
		t.Error("expected Blockers section")
	}
	if !strings.Contains(result, "- b1a [open] Blocker One") {
		t.Error("expected blocker line with status")
	}
	if !strings.Contains(result, "## Blocking") {
		t.Error("expected Blocking section")
	}
	if !strings.Contains(result, "- b2a [in_progress] Blocked By Me") {
		t.Error("expected blocking line with status")
	}
	if !strings.Contains(result, "## Children") {
		t.Error("expected Children section")
	}
	if !strings.Contains(result, "- c1a Child Task") {
		t.Error("expected child line without status (empty)")
	}
	if !strings.Contains(result, "## Linked") {
		t.Error("expected Linked section")
	}
	if !strings.Contains(result, "- l1a [open] Related") {
		t.Error("expected linked line with status")
	}
}

func TestFormatRelations_OnlyBlockers(t *testing.T) {
	rel := &TicketRelations{
		Blockers: []*Ticket{
			{ID: "aaa", Title: "Dep", Status: "open"},
		},
	}

	result := FormatRelations(rel)

	if !strings.Contains(result, "## Blockers") {
		t.Error("expected Blockers section")
	}
	if strings.Contains(result, "## Blocking") {
		t.Error("unexpected Blocking section")
	}
	if strings.Contains(result, "## Children") {
		t.Error("unexpected Children section")
	}
	if strings.Contains(result, "## Linked") {
		t.Error("unexpected Linked section")
	}
}

func TestFormatParentLine(t *testing.T) {
	rel := &TicketRelations{
		ParentTicket: &Ticket{ID: "par", Title: "Epic"},
	}

	result := FormatParentLine(rel)

	if result != "par (Epic)" {
		t.Errorf("expected 'par (Epic)', got %q", result)
	}
}

func TestFormatParentLine_NoParent(t *testing.T) {
	rel := &TicketRelations{}

	result := FormatParentLine(rel)

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}
