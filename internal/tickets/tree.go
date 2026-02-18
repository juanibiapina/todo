package tickets

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// treeNode represents a node in the dependency tree.
type treeNode struct {
	ticket   *Ticket
	children []*treeNode
	marker   string // "", "(cycle)", or "(dup)"
}

// DepTree generates a dependency tree string for the given ticket ID.
// If full is true, deduplication is disabled (same ticket can appear multiple times).
func DepTree(dir string, id string, full bool) (string, error) {
	// Resolve the ticket ID (supports partial matching)
	path, err := findTicketFile(dir, id)
	if err != nil {
		return "", err
	}
	resolvedID := strings.TrimSuffix(filepath.Base(path), ".md")

	// Load all tickets and build a lookup map
	allTickets, err := List(dir)
	if err != nil {
		return "", err
	}
	ticketMap := make(map[string]*Ticket)
	for _, t := range allTickets {
		ticketMap[t.ID] = t
	}

	root, ok := ticketMap[resolvedID]
	if !ok {
		return "", fmt.Errorf("ticket not found: %s", id)
	}

	// Build tree with cycle and dedup tracking
	ancestors := make(map[string]bool)
	visited := make(map[string]bool)
	rootNode := buildTreeNode(root, ticketMap, ancestors, visited, full)

	return formatTree(rootNode), nil
}

// buildTreeNode recursively builds a tree node from a ticket.
// ancestors tracks the current path for cycle detection.
// visited tracks all expanded nodes for dedup (when full is false).
func buildTreeNode(t *Ticket, ticketMap map[string]*Ticket, ancestors, visited map[string]bool, full bool) *treeNode {
	node := &treeNode{ticket: t}

	// Cycle detection: this ticket is an ancestor in the current path
	if ancestors[t.ID] {
		node.marker = "(cycle)"
		return node
	}

	// Dedup: this ticket was already expanded elsewhere (default mode only)
	if !full && visited[t.ID] {
		node.marker = "(dup)"
		return node
	}

	// Mark as ancestor (for cycle detection) and visited (for dedup)
	ancestors[t.ID] = true
	visited[t.ID] = true
	defer func() { delete(ancestors, t.ID) }()

	// Build children from deps
	var children []*treeNode
	for _, depID := range t.Deps {
		depTicket, ok := ticketMap[depID]
		if !ok {
			continue // skip missing deps
		}
		child := buildTreeNode(depTicket, ticketMap, ancestors, visited, full)
		children = append(children, child)
	}

	// Sort children: subtree depth descending, then ID ascending
	sort.Slice(children, func(i, j int) bool {
		di := subtreeDepth(children[i])
		dj := subtreeDepth(children[j])
		if di != dj {
			return di > dj
		}
		return children[i].ticket.ID < children[j].ticket.ID
	})

	node.children = children
	return node
}

// subtreeDepth returns the maximum depth of a node's subtree.
// Nodes with markers (cycle/dup) or no children have depth 0.
func subtreeDepth(n *treeNode) int {
	if n.marker != "" || len(n.children) == 0 {
		return 0
	}
	maxDepth := 0
	for _, child := range n.children {
		d := subtreeDepth(child)
		if d > maxDepth {
			maxDepth = d
		}
	}
	return maxDepth + 1
}

// formatTree renders a tree node and its children with box-drawing characters.
// Returns the formatted string without a trailing newline.
func formatTree(root *treeNode) string {
	var b strings.Builder
	b.WriteString(formatNodeLine(root))
	formatChildren(&b, root.children, "")
	return strings.TrimRight(b.String(), "\n")
}

// formatChildren recursively renders child nodes with appropriate prefixes.
func formatChildren(b *strings.Builder, children []*treeNode, prefix string) {
	for i, child := range children {
		isLast := i == len(children)-1

		connector := "├── "
		if isLast {
			connector = "└── "
		}

		b.WriteString(prefix)
		b.WriteString(connector)
		b.WriteString(formatNodeLine(child))

		// Only recurse into children if no marker (cycle/dup stops expansion)
		if child.marker == "" {
			childPrefix := prefix + "│   "
			if isLast {
				childPrefix = prefix + "    "
			}
			formatChildren(b, child.children, childPrefix)
		}
	}
}

// formatNodeLine renders a single node as "id [status] Title [marker]\n".
func formatNodeLine(n *treeNode) string {
	var parts []string
	parts = append(parts, n.ticket.ID)
	if n.ticket.Status != "" {
		parts = append(parts, fmt.Sprintf("[%s]", n.ticket.Status))
	}
	parts = append(parts, n.ticket.Title)
	if n.marker != "" {
		parts = append(parts, n.marker)
	}
	return strings.Join(parts, " ") + "\n"
}
