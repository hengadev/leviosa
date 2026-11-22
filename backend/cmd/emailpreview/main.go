// Command emailpreview renders every email of the preview catalogue
// (internal/notification/application/emailpreview) to HTML files with its
// fixed sample data, so the operator can send them to the client for review.
// Offline: no database, KMS, SMTP or network; nothing is sent.
package main

import (
	"flag"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/Leviosa-care/leviosa/backend/internal/notification/application/emailpreview"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
)

// preview is one file to write: an entry of the catalogue in one locale.
type preview struct {
	File   string
	Entry  emailpreview.Entry
	Locale domain.Locale
}

// Run writes every email of the preview catalogue, in each of its locales,
// to outDir and returns the path of the index linking them.
func Run(outDir string) (string, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	indexPath := filepath.Join(outDir, "index.html")
	var rows []string
	for _, p := range allPreviews() {
		rendered, err := emailpreview.Render(p.Entry.ID, p.Locale)
		if err != nil {
			return "", fmt.Errorf("render %s: %w", p.File, err)
		}
		if err := os.WriteFile(filepath.Join(outDir, p.File), []byte(rendered.HTML), 0o644); err != nil {
			return "", fmt.Errorf("write %s: %w", p.File, err)
		}
		rows = append(rows, indexRow(p, rendered))
	}
	if err := os.WriteFile(indexPath, []byte(indexHTML(rows)), 0o644); err != nil {
		return "", fmt.Errorf("write index.html: %w", err)
	}
	return indexPath, nil
}

// allPreviews lists one file per catalogue entry and locale:
// <id>_<locale>.html.
func allPreviews() []preview {
	var previews []preview
	for _, entry := range emailpreview.List() {
		for _, locale := range entry.Locales {
			previews = append(previews, preview{
				File:   entry.ID + "_" + string(locale) + ".html",
				Entry:  entry,
				Locale: locale,
			})
		}
	}
	return previews
}

func indexRow(p preview, rendered emailpreview.Rendered) string {
	attachments := ""
	if len(rendered.Attachments) > 0 {
		names := make([]string, 0, len(rendered.Attachments))
		for _, name := range rendered.Attachments {
			names = append(names, html.EscapeString(name))
		}
		attachments = fmt.Sprintf(`<div style="color:#666;">Joint&nbsp;: %s</div>`, strings.Join(names, ", "))
	}
	file := html.EscapeString(p.File)
	return fmt.Sprintf(`<li style="margin-bottom:14px;"><a href="%s">%s</a> (%s) — %s%s</li>`,
		file, html.EscapeString(p.Entry.Label), html.EscapeString(string(p.Locale)), html.EscapeString(rendered.Subject), attachments)
}

func indexHTML(rows []string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
  <head>
    <meta charset="UTF-8" />
    <title>Aperçu des emails — Leviosa</title>
  </head>
  <body style="font-family: Arial, sans-serif; max-width: 720px; margin: 40px auto;">
    <h1>Aperçu des emails de réservation</h1>
    <p>Données d'exemple (seed) — Camille Martin, %s, %s. Rien n'est envoyé.</p>
    <ul style="list-style: none; padding: 0;">
      %s
    </ul>
  </body>
</html>
`, emailpreview.SampleProduct, emailpreview.SampleAddress, strings.Join(rows, "\n      "))
}

func main() {
	outDir := flag.String("out", "tmp/email-preview", "output directory (created if missing)")
	flag.Parse()

	indexPath, err := Run(*outDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Email previews written to %s\nOpen %s\n", *outDir, indexPath)
}
