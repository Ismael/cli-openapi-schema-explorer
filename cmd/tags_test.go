package cmd

import (
	"strings"
	"testing"
)

func TestTags_ListWithDescriptions(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("tags-test.json"), "tags")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Users") {
		t.Error("expected 'Users' tag")
	}
	if !strings.Contains(out, "User management") {
		t.Error("expected Users description")
	}
	if !strings.Contains(out, "Admin") {
		t.Error("expected 'Admin' tag")
	}
}

func TestTags_TagWithoutDescription(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("tags-test.json"), "tags")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(out, "\n")
	found := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "Admin" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'Admin' tag without description on its own line")
	}
}

func TestTags_EmptySpec(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("empty-api.json"), "tags")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "#") {
		t.Error("expected hint comment even for empty tags")
	}
}

func TestTags_HintReferencesFilter(t *testing.T) {
	out, err := executeCommand("--spec", testdataPath("tags-test.json"), "tags")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--filter") {
		t.Error("expected hint to reference --filter")
	}
}
