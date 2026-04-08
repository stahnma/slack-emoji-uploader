package walker

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "partyparrot.gif"), []byte("gif"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "thumbsup.png"), []byte("png"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("readme"), 0644); err != nil {
		t.Fatal(err)
	}

	sub := filepath.Join(dir, "cats")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "Cat Wave.png"), []byte("png"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "cat-thumbsup.jpg"), []byte("jpg"), 0644); err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestWalkDir(t *testing.T) {
	dir := setupTestDir(t)
	entries, err := WalkDir(dir)
	if err != nil {
		t.Fatalf("WalkDir error: %v", err)
	}
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}
}

func TestEmojiName(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"partyparrot.gif", "partyparrot"},
		{"Cat Wave.png", "cat-wave"},
		{"LOUD.JPG", "loud"},
		{"my_emoji.jpeg", "my_emoji"},
	}
	for _, tt := range tests {
		got := EmojiName(tt.filename)
		if got != tt.want {
			t.Errorf("EmojiName(%q) = %q, want %q", tt.filename, got, tt.want)
		}
	}
}
