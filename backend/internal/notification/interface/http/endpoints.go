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

const (
	// PushStatusEndpoint says whether practitioner notifications (Web Push)
	// are available, with the VAPID public key (partner or administrator).
	PushStatusEndpoint = "/notifications/push"
	// PushSubscriptionEndpoint subscribes (PUT) or unsubscribes (DELETE) the
	// current device (partner or administrator).
	PushSubscriptionEndpoint = "/notifications/push/subscription"
)
