package tickets

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Ticket represents a single ticket.
type Ticket struct {
	Title       string
	ID          string
	Description string
	Status      string
	Type        string
	Priority    int
	Assignee    string
	Created     string
	Parent      string
	ExternalRef string
	Deps        []string
	Links       []string
	Tags        []string
}

// frontmatter is a helper struct for YAML marshaling of ticket metadata.
// It excludes Title and Description which are rendered outside the frontmatter.
type frontmatter struct {
	ID          string   `yaml:"id"`
	Status      string   `yaml:"status,omitempty"`
	Type        string   `yaml:"type,omitempty"`
	Priority    int      `yaml:"priority,omitempty"`
	Assignee    string   `yaml:"assignee,omitempty"`
	Created     string   `yaml:"created,omitempty"`
	Parent      string   `yaml:"parent,omitempty"`
	ExternalRef string   `yaml:"external_ref,omitempty"`
	Deps        []string `yaml:"deps,omitempty"`
	Links       []string `yaml:"links,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
}

// String returns a formatted single-line representation: "ID Title"
func (t *Ticket) String() string {
	return fmt.Sprintf("%s %s", t.ID, t.Title)
}

// FullString returns the full markdown representation of a ticket
// in YAML-frontmatter-first format.
func (t *Ticket) FullString() string {
	fm := frontmatter{
		ID:          t.ID,
		Status:      t.Status,
		Type:        t.Type,
		Priority:    t.Priority,
		Assignee:    t.Assignee,
		Created:     t.Created,
		Parent:      t.Parent,
		ExternalRef: t.ExternalRef,
		Deps:        t.Deps,
		Links:       t.Links,
		Tags:        t.Tags,
	}

	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		// Fallback: should never happen with simple types
		yamlBytes = []byte(fmt.Sprintf("id: %s\n", t.ID))
	}

	// yaml.Marshal adds a trailing newline; trim it since we add our own
	yamlStr := strings.TrimRight(string(yamlBytes), "\n")

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(yamlStr)
	b.WriteString("\n---\n")
	b.WriteString("# ")
	b.WriteString(t.Title)
	b.WriteString("\n")

	if t.Description != "" {
		b.WriteString(t.Description)
		b.WriteString("\n")
	}

	return b.String()
}
