package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRunWritesEveryPreview runs the command into a temp dir and checks
// that every expected file exists and is non-empty.
func TestRunWritesEveryPreview(t *testing.T) {
	outDir := t.TempDir()

	indexPath, err := Run(outDir)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	var files []string
	for _, p := range allPreviews() {
		files = append(files, p.File)
	}
	files = append(files, filepath.Base(indexPath))

	for _, name := range files {
		path := filepath.Join(outDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("expected file %s: %v", name, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("file %s is empty", name)
		}
	}
}
