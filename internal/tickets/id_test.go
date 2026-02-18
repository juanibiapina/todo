package tickets

import (
	"strings"
	"testing"
)

func TestGenerateID_Length(t *testing.T) {
	for i := 0; i < 100; i++ {
		id := generateID()
		if len(id) != 3 {
			t.Errorf("expected ID length 3, got %d: %q", len(id), id)
		}
	}
}

func TestGenerateID_Base62Chars(t *testing.T) {
	for i := 0; i < 100; i++ {
		id := generateID()
		for _, c := range id {
			if !strings.ContainsRune(base62Chars, c) {
				t.Errorf("ID %q contains invalid character %q", id, c)
			}
		}
	}
}

func TestGenerateID_Randomness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		seen[generateID()] = true
	}
	// With 62^3 = 238328 possible IDs, 100 draws should produce many unique values
	if len(seen) < 50 {
		t.Errorf("expected at least 50 unique IDs from 100 draws, got %d", len(seen))
	}
}

func TestGenerateUniqueID_EmptyMap(t *testing.T) {
	existing := make(map[string]bool)
	id := generateUniqueID(existing)
	if len(id) != 3 {
		t.Errorf("expected ID length 3, got %d: %q", len(id), id)
	}
	for _, c := range id {
		if !strings.ContainsRune(base62Chars, c) {
			t.Errorf("ID %q contains invalid character %q", id, c)
		}
	}
}

func TestGenerateUniqueID_AvoidsExisting(t *testing.T) {
	existing := map[string]bool{
		"abc": true,
		"XYZ": true,
		"123": true,
	}
	for i := 0; i < 100; i++ {
		id := generateUniqueID(existing)
		if existing[id] {
			t.Errorf("generateUniqueID returned existing ID %q", id)
		}
	}
}

func TestGenerateUniqueID_WithManyExisting(t *testing.T) {
	existing := make(map[string]bool)
	// Pre-populate with 1000 IDs to increase collision pressure
	for i := 0; i < 1000; i++ {
		existing[generateID()] = true
	}
	for i := 0; i < 100; i++ {
		id := generateUniqueID(existing)
		if existing[id] {
			t.Errorf("generateUniqueID returned existing ID %q", id)
		}
	}
}
