# Slack Emoji Uploader — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go CLI that bulk-uploads emoji images to free-tier Slack workspaces with idempotent state tracking and conflict resolution.

**Architecture:** Cobra CLI with three subcommands (`upload`, `resolve`, `status`). Internal packages for Slack HTTP client, state file management, and directory walking. Auth via session token + cookie from flags/env/.env file.

**Tech Stack:** Go, cobra, godotenv, standard library (net/http, encoding/json, filepath)

---

### Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `cmd/root.go`

**Step 1: Initialize Go module**

Run:
```bash
cd /Users/stahnma/go/src/github.com/stahnma/slack_emoji_uploader
go mod init github.com/stahnma/slack_emoji_uploader
```

**Step 2: Install dependencies**

Run:
```bash
go get github.com/spf13/cobra@latest
go get github.com/joho/godotenv@latest
```

**Step 3: Write `main.go`**

```go
package main

import (
	"os"

	"github.com/stahnma/slack_emoji_uploader/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

**Step 4: Write `cmd/root.go`**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	flagToken  string
	flagCookie string
	flagTeam   string
)

var rootCmd = &cobra.Command{
	Use:   "slack-emoji-uploader",
	Short: "Bulk upload emoji to Slack workspaces",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "Slack xoxc-* session token")
	rootCmd.PersistentFlags().StringVar(&flagCookie, "cookie", "", "Slack d= session cookie")
	rootCmd.PersistentFlags().StringVar(&flagTeam, "team", "", "Slack workspace subdomain")
}

func initConfig() {
	_ = godotenv.Load() // best-effort .env loading
}

// resolveAuth returns token, cookie, team with precedence: flags > env > .env
func resolveAuth() (token, cookie, team string, err error) {
	token = flagToken
	if token == "" {
		token = os.Getenv("SLACK_TOKEN")
	}
	cookie = flagCookie
	if cookie == "" {
		cookie = os.Getenv("SLACK_COOKIE")
	}
	team = flagTeam
	if team == "" {
		team = os.Getenv("SLACK_TEAM")
	}
	if token == "" || cookie == "" || team == "" {
		return "", "", "", fmt.Errorf("missing required auth: token, cookie, and team must be set via flags, env vars, or .env file")
	}
	return token, cookie, team, nil
}
```

**Step 5: Verify it compiles**

Run:
```bash
go build ./...
```
Expected: no errors

**Step 6: Commit**

```bash
git add main.go cmd/root.go go.mod go.sum
git commit -m "feat: project scaffolding with cobra CLI and auth loading"
```

---

### Task 2: Directory Walker

**Files:**
- Create: `internal/walker/walker.go`
- Create: `internal/walker/walker_test.go`

**Step 1: Write failing tests**

```go
package walker

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create files
	os.WriteFile(filepath.Join(dir, "partyparrot.gif"), []byte("gif"), 0644)
	os.WriteFile(filepath.Join(dir, "thumbsup.png"), []byte("png"), 0644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("readme"), 0644)

	// Create subdirectory with files
	sub := filepath.Join(dir, "cats")
	os.Mkdir(sub, 0755)
	os.WriteFile(filepath.Join(sub, "Cat Wave.png"), []byte("png"), 0644)
	os.WriteFile(filepath.Join(sub, "cat-thumbsup.jpg"), []byte("jpg"), 0644)

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
```

**Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/walker/ -v
```
Expected: FAIL — functions not defined

**Step 3: Write implementation**

```go
package walker

import (
	"os"
	"path/filepath"
	"strings"
)

var supportedExts = map[string]bool{
	".png":  true,
	".gif":  true,
	".jpg":  true,
	".jpeg": true,
}

// Entry represents a discovered emoji file.
type Entry struct {
	Path string // full path to the file
	Name string // derived emoji name
}

// WalkDir recursively finds all supported image files in dir.
func WalkDir(dir string) ([]Entry, error) {
	var entries []Entry
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !supportedExts[ext] {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		entries = append(entries, Entry{
			Path: rel,
			Name: EmojiName(info.Name()),
		})
		return nil
	})
	return entries, err
}

// EmojiName derives an emoji name from a filename.
// Strips extension, lowercases, replaces spaces with hyphens.
func EmojiName(filename string) string {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	return name
}
```

**Step 4: Run tests to verify they pass**

Run:
```bash
go test ./internal/walker/ -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/walker/
git commit -m "feat: directory walker with emoji name derivation"
```

---

### Task 3: State Management

**Files:**
- Create: `internal/state/state.go`
- Create: `internal/state/state_test.go`

**Step 1: Write failing tests**

```go
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
```

**Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/state/ -v
```
Expected: FAIL

**Step 3: Write implementation**

```go
package state

import (
	"encoding/json"
	"os"
	"time"
)

// StateEntry records a successful upload.
type StateEntry struct {
	Name       string    `json:"name"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// State tracks uploaded emoji.
type State struct {
	Entries map[string]StateEntry `json:"entries"`
	path    string
}

// Load reads state from path, or returns empty state if file doesn't exist.
func Load(path string) (*State, error) {
	s := &State{
		Entries: make(map[string]StateEntry),
		path:    path,
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &s.Entries); err != nil {
		return nil, err
	}
	return s, nil
}

// Save writes state to disk.
func (s *State) Save() error {
	data, err := json.MarshalIndent(s.Entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

// RecordSuccess marks a file as successfully uploaded.
func (s *State) RecordSuccess(filePath, emojiName string) {
	s.Entries[filePath] = StateEntry{
		Name:       emojiName,
		UploadedAt: time.Now().UTC(),
	}
}

// IsUploaded returns true if the file has been uploaded.
func (s *State) IsUploaded(filePath string) bool {
	_, ok := s.Entries[filePath]
	return ok
}

// ConflictEntry records a name conflict.
type ConflictEntry struct {
	Name        string    `json:"name"`
	Error       string    `json:"error"`
	Attempted   []string  `json:"attempted"`
	LastAttempt time.Time `json:"last_attempt"`
}

// Conflicts tracks emoji with name collisions.
type Conflicts struct {
	Entries map[string]ConflictEntry `json:"entries"`
	path    string
}

// LoadConflicts reads conflicts from path, or returns empty if file doesn't exist.
func LoadConflicts(path string) (*Conflicts, error) {
	c := &Conflicts{
		Entries: make(map[string]ConflictEntry),
		path:    path,
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &c.Entries); err != nil {
		return nil, err
	}
	return c, nil
}

// Save writes conflicts to disk.
func (c *Conflicts) Save() error {
	data, err := json.MarshalIndent(c.Entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0644)
}

// RecordConflict logs a name conflict for a file.
func (c *Conflicts) RecordConflict(filePath, emojiName string, attempted []string) {
	c.Entries[filePath] = ConflictEntry{
		Name:        emojiName,
		Error:       "error_name_taken",
		Attempted:   attempted,
		LastAttempt: time.Now().UTC(),
	}
}

// Remove deletes a conflict entry (used after successful resolution).
func (c *Conflicts) Remove(filePath string) {
	delete(c.Entries, filePath)
}
```

**Step 4: Run tests to verify they pass**

Run:
```bash
go test ./internal/state/ -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/state/
git commit -m "feat: state and conflict tracking with JSON persistence"
```

---

### Task 4: Slack HTTP Client

**Files:**
- Create: `internal/slack/client.go`
- Create: `internal/slack/client_test.go`

**Step 1: Write failing tests**

```go
package slack

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUploadEmoji_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/emoji.add" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		// Verify multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.FormValue("name") != "partyparrot" {
			t.Errorf("name = %q", r.FormValue("name"))
		}
		if r.FormValue("token") == "" {
			t.Error("missing token")
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	}))
	defer server.Close()

	c := NewClient("xoxc-test", "d=test", "testteam", 0)
	c.baseURL = server.URL

	result, err := c.UploadEmoji("partyparrot", []byte("fake-image"), "partyparrot.gif")
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if !result.OK {
		t.Error("expected ok=true")
	}
}

func TestUploadEmoji_NameTaken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": "error_name_taken",
		})
	}))
	defer server.Close()

	c := NewClient("xoxc-test", "d=test", "testteam", 0)
	c.baseURL = server.URL

	result, err := c.UploadEmoji("partyparrot", []byte("fake-image"), "partyparrot.gif")
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if result.OK {
		t.Error("expected ok=false")
	}
	if result.Error != "error_name_taken" {
		t.Errorf("error = %q", result.Error)
	}
}

func TestUploadEmoji_RateLimit(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.Header().Set("Retry-After", "1")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":    false,
				"error": "ratelimited",
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	}))
	defer server.Close()

	c := NewClient("xoxc-test", "d=test", "testteam", 0)
	c.baseURL = server.URL
	c.baseBackoff = 10 * time.Millisecond // fast backoff for tests

	result, err := c.UploadEmoji("partyparrot", []byte("fake-image"), "partyparrot.gif")
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if !result.OK {
		t.Error("expected eventual success")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}
```

**Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/slack/ -v
```
Expected: FAIL

**Step 3: Write implementation**

```go
package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"
)

// UploadResult represents the Slack API response.
type UploadResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// Client handles Slack emoji API calls.
type Client struct {
	token       string
	cookie      string
	team        string
	delay       time.Duration
	baseURL     string // overridden in tests
	baseBackoff time.Duration
	httpClient  *http.Client
}

// NewClient creates a Slack client.
func NewClient(token, cookie, team string, delay time.Duration) *Client {
	return &Client{
		token:       token,
		cookie:      cookie,
		team:        team,
		delay:       delay,
		baseURL:     fmt.Sprintf("https://%s.slack.com", team),
		baseBackoff: 2 * time.Second,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// UploadEmoji uploads an emoji image to Slack.
// It retries on rate limits with exponential backoff.
func (c *Client) UploadEmoji(name string, imageData []byte, filename string) (*UploadResult, error) {
	maxRetries := 5
	backoff := c.baseBackoff

	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, err := c.doUpload(name, imageData, filename)
		if err != nil {
			return nil, err
		}
		if result.Error != "ratelimited" {
			return result, nil
		}
		if attempt == maxRetries {
			return result, nil
		}
		time.Sleep(backoff)
		backoff *= 2
		if backoff > 60*time.Second {
			backoff = 60 * time.Second
		}
	}
	return nil, fmt.Errorf("unreachable")
}

func (c *Client) doUpload(name string, imageData []byte, filename string) (*UploadResult, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	writer.WriteField("name", name)
	writer.WriteField("token", c.token)
	writer.WriteField("mode", "data")

	ext := filepath.Ext(filename)
	contentType := "image/png"
	switch ext {
	case ".gif":
		contentType = "image/gif"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	}

	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(imageData)); err != nil {
		return nil, fmt.Errorf("write image: %w", err)
	}
	_ = contentType // content type is set by CreateFormFile via filename extension
	writer.Close()

	req, err := http.NewRequest("POST", c.baseURL+"/api/emoji.add", &body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Cookie", "d="+c.cookie)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result UploadResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w (body: %s)", err, string(respBody))
	}
	return &result, nil
}

// Delay returns the configured delay between uploads.
func (c *Client) Delay() time.Duration {
	return c.delay
}
```

**Step 4: Run tests to verify they pass**

Run:
```bash
go test ./internal/slack/ -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/slack/
git commit -m "feat: Slack HTTP client with emoji upload and rate limit backoff"
```

---

### Task 5: Upload Subcommand

**Files:**
- Create: `cmd/upload.go`

**Step 1: Write the upload command**

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/stahnma/slack_emoji_uploader/internal/slack"
	"github.com/stahnma/slack_emoji_uploader/internal/state"
	"github.com/stahnma/slack_emoji_uploader/internal/walker"
)

var (
	flagAutoSuffix bool
	flagDelay      time.Duration
	flagDryRun     bool
)

var uploadCmd = &cobra.Command{
	Use:   "upload <directory>",
	Short: "Upload emoji from a directory to Slack",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpload,
}

func init() {
	uploadCmd.Flags().BoolVar(&flagAutoSuffix, "auto-suffix", false, "Automatically append numeric suffix on name conflicts")
	uploadCmd.Flags().DurationVar(&flagDelay, "delay", 1*time.Second, "Delay between uploads")
	uploadCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Show what would be uploaded without uploading")
	rootCmd.AddCommand(uploadCmd)
}

func runUpload(cmd *cobra.Command, args []string) error {
	dir := args[0]

	token, cookie, team, err := resolveAuth()
	if err != nil {
		return err
	}

	entries, err := walker.WalkDir(dir)
	if err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}
	if len(entries) == 0 {
		fmt.Println("No emoji files found.")
		return nil
	}

	st, err := state.Load("emoji-state.json")
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}
	conflicts, err := state.LoadConflicts("emoji-conflicts.json")
	if err != nil {
		return fmt.Errorf("loading conflicts: %w", err)
	}

	client := slack.NewClient(token, cookie, team, flagDelay)

	var uploaded, skipped, conflicted int

	for _, entry := range entries {
		if st.IsUploaded(entry.Path) {
			skipped++
			continue
		}

		if flagDryRun {
			fmt.Printf("[dry-run] Would upload: %s → :%s:\n", entry.Path, entry.Name)
			continue
		}

		name := entry.Name
		attempted := []string{}
		success := false

		for suffix := 0; suffix < 100; suffix++ {
			candidateName := name
			if suffix > 0 {
				candidateName = fmt.Sprintf("%s%d", name, suffix+1)
			}
			attempted = append(attempted, candidateName)

			imageData, err := os.ReadFile(filepath.Join(dir, entry.Path))
			if err != nil {
				return fmt.Errorf("reading %s: %w", entry.Path, err)
			}

			fmt.Printf("Uploading: %s → :%s: ... ", entry.Path, candidateName)
			result, err := client.UploadEmoji(candidateName, imageData, filepath.Base(entry.Path))
			if err != nil {
				return fmt.Errorf("uploading %s: %w", entry.Path, err)
			}

			if result.OK {
				fmt.Println("OK")
				st.RecordSuccess(entry.Path, candidateName)
				if err := st.Save(); err != nil {
					return fmt.Errorf("saving state: %w", err)
				}
				uploaded++
				success = true
				break
			}

			if result.Error == "error_name_taken" {
				fmt.Printf("conflict (%s)\n", candidateName)
				if !flagAutoSuffix {
					break
				}
				continue
			}

			if result.Error == "not_authed" || result.Error == "invalid_auth" || result.Error == "token_revoked" {
				return fmt.Errorf("authentication failed: %s — check your token and cookie", result.Error)
			}

			return fmt.Errorf("unexpected error uploading %s: %s", entry.Path, result.Error)
		}

		if !success {
			conflicts.RecordConflict(entry.Path, name, attempted)
			if err := conflicts.Save(); err != nil {
				return fmt.Errorf("saving conflicts: %w", err)
			}
			conflicted++
		}

		time.Sleep(client.Delay())
	}

	fmt.Printf("\nDone: %d uploaded, %d skipped, %d conflicts\n", uploaded, skipped, conflicted)
	return nil
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./...
```
Expected: no errors

**Step 3: Commit**

```bash
git add cmd/upload.go
git commit -m "feat: upload subcommand with idempotency and auto-suffix"
```

---

### Task 6: Status Subcommand

**Files:**
- Create: `cmd/status.go`

**Step 1: Write the status command**

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stahnma/slack_emoji_uploader/internal/state"
	"github.com/stahnma/slack_emoji_uploader/internal/walker"
)

var statusCmd = &cobra.Command{
	Use:   "status <directory>",
	Short: "Show upload progress",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	dir := args[0]

	entries, err := walker.WalkDir(dir)
	if err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	st, err := state.Load("emoji-state.json")
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}
	conflicts, err := state.LoadConflicts("emoji-conflicts.json")
	if err != nil {
		return fmt.Errorf("loading conflicts: %w", err)
	}

	total := len(entries)
	uploaded := len(st.Entries)
	conflicted := len(conflicts.Entries)
	remaining := total - uploaded - conflicted

	fmt.Printf("Total:      %d\n", total)
	fmt.Printf("Uploaded:   %d\n", uploaded)
	fmt.Printf("Conflicts:  %d\n", conflicted)
	fmt.Printf("Remaining:  %d\n", remaining)

	return nil
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./...
```
Expected: no errors

**Step 3: Commit**

```bash
git add cmd/status.go
git commit -m "feat: status subcommand showing upload progress"
```

---

### Task 7: Resolve Subcommand

**Files:**
- Create: `cmd/resolve.go`

**Step 1: Write the resolve command**

```go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/stahnma/slack_emoji_uploader/internal/slack"
	"github.com/stahnma/slack_emoji_uploader/internal/state"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve <directory>",
	Short: "Interactively resolve emoji name conflicts",
	Args:  cobra.ExactArgs(1),
	RunE:  runResolve,
}

func init() {
	rootCmd.AddCommand(resolveCmd)
}

func runResolve(cmd *cobra.Command, args []string) error {
	dir := args[0]

	token, cookie, team, err := resolveAuth()
	if err != nil {
		return err
	}

	st, err := state.Load("emoji-state.json")
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}
	conflicts, err := state.LoadConflicts("emoji-conflicts.json")
	if err != nil {
		return fmt.Errorf("loading conflicts: %w", err)
	}

	if len(conflicts.Entries) == 0 {
		fmt.Println("No conflicts to resolve.")
		return nil
	}

	client := slack.NewClient(token, cookie, team, 1*time.Second)
	reader := bufio.NewReader(os.Stdin)
	resolved := 0

	for filePath, entry := range conflicts.Entries {
		fmt.Printf("\nConflict: %s → :%s: already exists\n", filePath, entry.Name)
		fmt.Printf("Attempted: %s\n", strings.Join(entry.Attempted, ", "))
		fmt.Print("Enter new name (or 'skip'): ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" || input == "skip" {
			fmt.Println("Skipped.")
			continue
		}

		imageData, err := os.ReadFile(filepath.Join(dir, filePath))
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			continue
		}

		fmt.Printf("Uploading: %s → :%s: ... ", filePath, input)
		result, err := client.UploadEmoji(input, imageData, filepath.Base(filePath))
		if err != nil {
			fmt.Printf("error: %v\n", err)
			continue
		}

		if result.OK {
			fmt.Println("OK")
			st.RecordSuccess(filePath, input)
			conflicts.Remove(filePath)
			resolved++
		} else {
			fmt.Printf("failed: %s\n", result.Error)
			entry.Attempted = append(entry.Attempted, input)
			entry.LastAttempt = time.Now().UTC()
			conflicts.Entries[filePath] = entry
		}

		st.Save()
		conflicts.Save()
		time.Sleep(client.Delay())
	}

	fmt.Printf("\nResolved: %d, Remaining: %d\n", resolved, len(conflicts.Entries))
	return nil
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./...
```
Expected: no errors

**Step 3: Commit**

```bash
git add cmd/resolve.go
git commit -m "feat: resolve subcommand for interactive conflict resolution"
```

---

### Task 8: README with Credential Extraction Guide

**Files:**
- Create: `README.md`

**Step 1: Write the README**

Write a README.md covering:
- Project description and features
- Installation (`go install`)
- **Getting Your Slack Credentials** — detailed step-by-step:
  1. Open workspace in browser at `https://your-team.slack.com`
  2. Open DevTools (F12 / Cmd+Option+I)
  3. Go to Network tab
  4. Navigate to `https://your-team.slack.com/customize/emoji`
  5. In the Network tab, find a request to any `/api/` endpoint
  6. In the request payload or headers, find the `xoxc-*` token value
  7. Go to Application tab → Cookies → find the `d` cookie value
  8. Create a `.env` file with `SLACK_TOKEN`, `SLACK_COOKIE`, `SLACK_TEAM`
- Usage examples for all three subcommands
- `.env` file format
- Flag reference
- Add `.env` to `.gitignore` example

**Step 2: Create `.gitignore`**

```
.env
emoji-state.json
emoji-conflicts.json
```

**Step 3: Commit**

```bash
git add README.md .gitignore
git commit -m "docs: README with credential extraction guide and .gitignore"
```

---

### Task 9: Final Integration Test

**Step 1: Build the binary**

Run:
```bash
go build -o slack-emoji-uploader .
```
Expected: produces binary

**Step 2: Test help output**

Run:
```bash
./slack-emoji-uploader --help
./slack-emoji-uploader upload --help
./slack-emoji-uploader status --help
./slack-emoji-uploader resolve --help
```
Expected: all show usage info

**Step 3: Test dry-run with a test directory**

Run:
```bash
mkdir -p /tmp/test-emoji/cats
cp some-test-images or create dummy files
./slack-emoji-uploader upload --dry-run --team test --token fake --cookie fake /tmp/test-emoji/
```
Expected: lists files that would be uploaded

**Step 4: Run all tests**

Run:
```bash
go test ./... -v
```
Expected: all pass

**Step 5: Commit any fixes**

```bash
git add -A
git commit -m "chore: final integration verification"
```

---

Plan complete and saved to `docs/plans/2026-04-07-implementation-plan.md`. Two execution options:

**1. Subagent-Driven (this session)** — I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** — Open a new session with executing-plans, batch execution with checkpoints

Which approach?