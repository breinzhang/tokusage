package claude

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func DiscoverJSONLFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".jsonl") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
