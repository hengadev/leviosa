package domain

import (
	"bytes"
	"fmt"
	"html/template"
)

// RenderEmailTemplate executes the embedded email template
// templates/<name>.html with data. The SMTP client and the email preview
// catalogue both render through it, so a preview is exactly what is sent.
func RenderEmailTemplate(name string, data any) (string, error) {
	tmpl, err := template.ParseFS(EmailTemplates, "templates/"+name+".html")
	if err != nil {
		return "", fmt.Errorf("parse email template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute email template %s: %w", name, err)
	}

	return buf.String(), nil
}
