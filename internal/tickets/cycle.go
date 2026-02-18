package tickets

import (
	"fmt"
	"sort"
	"strings"
)

// DepCycles detects dependency cycles among open (non-closed) tickets.
// Returns a formatted string of all cycles found, or empty string if none.
func DepCycles(dir string) (string, error) {
	allTickets, err := List(dir)
	if err != nil {
		return "", err
	}

	// Build ticket map and adjacency graph, excluding closed tickets
	ticketMap := make(map[string]*Ticket)
	for _, t := range allTickets {
		if t.Status != "closed" {
			ticketMap[t.ID] = t
		}
	}

	cycles := findCycles(ticketMap)

	if len(cycles) == 0 {
		return "", nil
	}

	return formatCycles(cycles, ticketMap), nil
}

// findCycles performs DFS-based cycle detection on the dependency graph.
// Uses 3-color marking: white (unseen), gray (in current path), black (fully explored).
// Returns deduplicated, normalized cycles.
func findCycles(ticketMap map[string]*Ticket) [][]string {
	const (
		white = 0
		gray  = 1
		black = 2
	)

	color := make(map[string]int)
	path := make([]string, 0)
	var rawCycles [][]string

	var dfs func(id string)
	dfs = func(id string) {
		color[id] = gray
		path = append(path, id)

		t, ok := ticketMap[id]
		if ok {
			for _, depID := range t.Deps {
				// Skip deps that point to closed or non-existent tickets
				if _, exists := ticketMap[depID]; !exists {
					continue
				}

				switch color[depID] {
				case white:
					dfs(depID)
				case gray:
					// Found a cycle — extract from the gray node to end of path
					cycle := extractCycle(path, depID)
					rawCycles = append(rawCycles, cycle)
				}
				// black: already fully explored, skip
			}
		}

		path = path[:len(path)-1]
		color[id] = black
	}

	// Sort IDs for deterministic traversal order
	var ids []string
	for id := range ticketMap {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		if color[id] == white {
			dfs(id)
		}
	}

	return deduplicateCycles(rawCycles)
}

// extractCycle extracts a cycle from the current DFS path.
// The cycle starts at startID and includes all nodes from there to the end of path.
func extractCycle(path []string, startID string) []string {
	for i, id := range path {
		if id == startID {
			cycle := make([]string, len(path)-i)
			copy(cycle, path[i:])
			return cycle
		}
	}
	return nil
}

// normalizeCycle rotates a cycle so that the lexicographically smallest ID is first.
func normalizeCycle(cycle []string) []string {
	if len(cycle) == 0 {
		return cycle
	}

	// Find the index of the smallest ID
	minIdx := 0
	for i := 1; i < len(cycle); i++ {
		if cycle[i] < cycle[minIdx] {
			minIdx = i
		}
	}

	// Rotate so smallest is first
	normalized := make([]string, len(cycle))
	for i := 0; i < len(cycle); i++ {
		normalized[i] = cycle[(minIdx+i)%len(cycle)]
	}
	return normalized
}

// deduplicateCycles normalizes all cycles and removes duplicates.
// Sorts the result by first element, then by length.
func deduplicateCycles(rawCycles [][]string) [][]string {
	seen := make(map[string]bool)
	var unique [][]string

	for _, cycle := range rawCycles {
		normalized := normalizeCycle(cycle)
		key := strings.Join(normalized, ",")
		if !seen[key] {
			seen[key] = true
			unique = append(unique, normalized)
		}
	}

	// Sort cycles: by first element, then by length
	sort.Slice(unique, func(i, j int) bool {
		if unique[i][0] != unique[j][0] {
			return unique[i][0] < unique[j][0]
		}
		return len(unique[i]) < len(unique[j])
	})

	return unique
}

// formatCycles renders all detected cycles as a formatted string.
func formatCycles(cycles [][]string, ticketMap map[string]*Ticket) string {
	var b strings.Builder

	for i, cycle := range cycles {
		if i > 0 {
			b.WriteString("\n")
		}

		// Header line: "Cycle: aaa -> bbb -> ccc -> aaa"
		b.WriteString("Cycle: ")
		for j, id := range cycle {
			if j > 0 {
				b.WriteString(" -> ")
			}
			b.WriteString(id)
		}
		b.WriteString(" -> ")
		b.WriteString(cycle[0])
		b.WriteString("\n")

		// Member details
		for _, id := range cycle {
			b.WriteString(formatCycleMember(id, ticketMap))
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

// formatCycleMember renders a single cycle member as "  id [status] Title\n".
func formatCycleMember(id string, ticketMap map[string]*Ticket) string {
	t, ok := ticketMap[id]
	if !ok {
		return fmt.Sprintf("  %s\n", id)
	}

	var parts []string
	parts = append(parts, id)
	if t.Status != "" {
		parts = append(parts, fmt.Sprintf("[%s]", t.Status))
	}
	parts = append(parts, t.Title)

	return fmt.Sprintf("  %s\n", strings.Join(parts, " "))
}
