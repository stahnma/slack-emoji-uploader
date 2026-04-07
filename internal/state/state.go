package state

import (
	"encoding/json"
	"os"
	"time"
)

type StateEntry struct {
	Name       string    `json:"name"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type State struct {
	Entries map[string]StateEntry `json:"entries"`
	path    string
}

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

func (s *State) Save() error {
	data, err := json.MarshalIndent(s.Entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *State) RecordSuccess(filePath, emojiName string) {
	s.Entries[filePath] = StateEntry{
		Name:       emojiName,
		UploadedAt: time.Now().UTC(),
	}
}

func (s *State) IsUploaded(filePath string) bool {
	_, ok := s.Entries[filePath]
	return ok
}

type ConflictEntry struct {
	Name        string    `json:"name"`
	Error       string    `json:"error"`
	Attempted   []string  `json:"attempted"`
	LastAttempt time.Time `json:"last_attempt"`
}

type Conflicts struct {
	Entries map[string]ConflictEntry `json:"entries"`
	path    string
}

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

func (c *Conflicts) Save() error {
	data, err := json.MarshalIndent(c.Entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0644)
}

func (c *Conflicts) RecordConflict(filePath, emojiName string, attempted []string) {
	c.Entries[filePath] = ConflictEntry{
		Name:        emojiName,
		Error:       "error_name_taken",
		Attempted:   attempted,
		LastAttempt: time.Now().UTC(),
	}
}

func (c *Conflicts) Remove(filePath string) {
	delete(c.Entries, filePath)
}
