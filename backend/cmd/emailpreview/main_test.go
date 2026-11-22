package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunWritesEveryPreview runs the command into a temp dir and checks
// that every catalogue entry is written in each of its locales (booking
// client emails FR and EN, partner notices FR, the other emails in the one
// language they are sent in), non-empty and linked from the index.
func TestRunWritesEveryPreview(t *testing.T) {
	outDir := t.TempDir()

	indexPath, err := Run(outDir)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	previews := allPreviews()
	if len(previews) != 25 {
		t.Fatalf("allPreviews returned %d files, want 25 (9 booking client emails x 2 locales + 2 partner notices + 5 other emails)", len(previews))
	}
	for _, name := range []string{"payment_receipt_en.html", "payment_failed_en.html", "contact_form_fr.html", "payment_dispute_alert_en.html"} {
		found := false
		for _, p := range previews {
			found = found || p.File == name
		}
		if !found {
			t.Errorf("allPreviews misses %s", name)
		}
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
