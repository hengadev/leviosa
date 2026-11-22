package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunWritesEveryPreview runs the command into a temp dir and checks
// that every catalogue entry is written in each of its locales (client
// emails FR and EN, partner notices FR), non-empty and linked from the index.
func TestRunWritesEveryPreview(t *testing.T) {
	outDir := t.TempDir()

	indexPath, err := Run(outDir)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	previews := allPreviews()
	if len(previews) != 20 {
		t.Fatalf("allPreviews returned %d files, want 20 (9 client emails x 2 locales + 2 partner notices)", len(previews))
	}

	index, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}

	var files []string
	for _, p := range previews {
		files = append(files, p.File)
		if !strings.Contains(string(index), `href="`+p.File+`"`) {
			t.Errorf("index does not link %s", p.File)
		}
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
