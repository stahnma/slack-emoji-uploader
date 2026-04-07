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

type Entry struct {
	Path string // relative path to the file from the walk root
	Name string // derived emoji name
}

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

func EmojiName(filename string) string {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	return name
}
