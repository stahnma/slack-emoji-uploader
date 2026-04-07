package state

import (
	"path/filepath"
	"testing"
)

func TestStateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "emoji-state.json")

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load new: %v", err)
	}
	if len(s.Entries) != 0 {
		t.Fatal("expected empty state")
	}

	s.RecordSuccess("emoji/partyparrot.gif", "partyparrot")
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	s2, err := Load(path)
	if err != nil {
		t.Fatalf("Load existing: %v", err)
	}
	entry, ok := s2.Entries["emoji/partyparrot.gif"]
	if !ok {
		t.Fatal("expected entry for partyparrot")
	}
	if entry.Name != "partyparrot" {
		t.Errorf("name = %q, want partyparrot", entry.Name)
	}
}

func TestStateIsUploaded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "emoji-state.json")

	s, _ := Load(path)
	if s.IsUploaded("emoji/partyparrot.gif") {
		t.Fatal("should not be uploaded yet")
	}
	s.RecordSuccess("emoji/partyparrot.gif", "partyparrot")
	if !s.IsUploaded("emoji/partyparrot.gif") {
		t.Fatal("should be uploaded")
	}
}

func TestConflictsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "emoji-conflicts.json")

	c, err := LoadConflicts(path)
	if err != nil {
		t.Fatalf("Load new: %v", err)
	}

	c.RecordConflict("emoji/partyparrot.gif", "partyparrot", []string{"partyparrot"})
	if err := c.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	c2, err := LoadConflicts(path)
	if err != nil {
		t.Fatalf("Load existing: %v", err)
	}
	entry, ok := c2.Entries["emoji/partyparrot.gif"]
	if !ok {
		t.Fatal("expected conflict entry")
	}
	if entry.Name != "partyparrot" {
		t.Errorf("name = %q, want partyparrot", entry.Name)
	}
}
