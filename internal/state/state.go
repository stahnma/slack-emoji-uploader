package state

import (
	"encoding/json"
	"errors"
	"io/fs"
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
	if errors.Is(err, fs.ErrNotExist) {
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

// Save writes state to disk atomically.
func (s *State) Save() error {
	return atomicWrite(s.path, s.Entries)
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
	if errors.Is(err, fs.ErrNotExist) {
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

// Save writes conflicts to disk atomically.
func (c *Conflicts) Save() error {
	return atomicWrite(c.path, c.Entries)
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

// atomicWrite writes data as indented JSON to path via a temp file + rename.
func atomicWrite(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
