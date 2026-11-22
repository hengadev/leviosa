package http

const (
	// ContactEndpoint receives public contact-form submissions (unauthenticated).
	ContactEndpoint = "/contact"

	// AdminEmailPreviewsEndpoint lists the previewable emails (administrator).
	AdminEmailPreviewsEndpoint = "/admin/emails/previews"
	// AdminEmailPreviewEndpoint renders one email with sample data
	// (administrator); ?locale=fr|en, French by default.
	AdminEmailPreviewEndpoint = "/admin/emails/previews/{id}"
)
