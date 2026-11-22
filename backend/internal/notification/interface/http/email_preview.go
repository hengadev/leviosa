package http

import (
	"net/http"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/application/emailpreview"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
)

// emailPreviewSummary is one entry of the email preview list.
type emailPreviewSummary struct {
	ID       string            `json:"id"`
	Label    string            `json:"label"`
	Audience string            `json:"audience"`
	Locales  []string          `json:"locales"`
	Subjects map[string]string `json:"subjects"`
}

// emailPreviewResponse is one email rendered with the sample data.
type emailPreviewResponse struct {
	Subject     string   `json:"subject"`
	HTML        string   `json:"html"`
	Attachments []string `json:"attachments"`
}

// ListEmailPreviews lists every previewable email with its subject in each
// of its locales.
func (h *handler) ListEmailPreviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	entries := emailpreview.List()
	summaries := make([]emailPreviewSummary, 0, len(entries))
	for _, e := range entries {
		summary := emailPreviewSummary{
			ID:       e.ID,
			Label:    e.Label,
			Audience: string(e.Audience),
			Locales:  make([]string, 0, len(e.Locales)),
			Subjects: make(map[string]string, len(e.Locales)),
		}
		for _, locale := range e.Locales {
			preview, err := emailpreview.Render(e.ID, locale)
			if err != nil {
				httpx.RespondWithServiceError(w, logger, ctx, err, "list email previews")
				return
			}
			summary.Locales = append(summary.Locales, string(locale))
			summary.Subjects[string(locale)] = preview.Subject
		}
		summaries = append(summaries, summary)
	}

	httpx.RespondWithJSON(w, summaries, http.StatusOK)
}

// GetEmailPreview renders one email with the sample data in the requested
// locale (French when none is given). An unknown id or a locale the email
// does not exist in is a 404.
func (h *handler) GetEmailPreview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	locale := domain.Locale(r.URL.Query().Get("locale"))
	if locale == "" {
		locale = domain.LocaleFR
	}

	preview, err := emailpreview.Render(r.PathValue("id"), locale)
	if err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "render email preview")
		return
	}

	httpx.RespondWithJSON(w, emailPreviewResponse{
		Subject:     preview.Subject,
		HTML:        preview.HTML,
		Attachments: preview.Attachments,
	}, http.StatusOK)
}
