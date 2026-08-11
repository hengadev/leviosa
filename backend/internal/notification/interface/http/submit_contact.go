package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/common/httpx"
	"github.com/Leviosa-care/leviosa/backend/internal/common/validation"
)

const (
	contactNameMaxLength    = 200
	contactMessageMaxLength = 5000
)

// contactFormRequest is the public contact-form submission payload.
type contactFormRequest struct {
	Name         string `json:"name"`
	ContactEmail string `json:"contact_email,omitempty"`
	ContactPhone string `json:"contact_phone,omitempty"`
	Message      string `json:"message"`
}

// SubmitContact handles unauthenticated contact-form submissions from the
// public storefront. It routes the message to staff via the existing
// notification-email channel rather than a new inbound pipeline (see
// docs/prd/public-trust-and-seo-foundations.md).
func (h *handler) SubmitContact(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		httpx.RespondWithError(w, errors.New("unsupported media type: please send 'application/json'"), http.StatusUnsupportedMediaType)
		return
	}

	ctx := r.Context()

	logger, err := ctxutil.GetLoggerFromContext(ctx)
	if err != nil {
		httpx.RespondWithError(w, err, http.StatusInternalServerError)
		return
	}

	var request contactFormRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		httpx.RespondWithError(w, errs.NewInvalidValueErr(fmt.Sprintf("invalid request body: %v", err)), http.StatusBadRequest)
		return
	}

	request.Name = strings.TrimSpace(request.Name)
	request.ContactEmail = strings.TrimSpace(request.ContactEmail)
	// Strip all whitespace (not just leading/trailing) from the phone number —
	// validation.ValidatePhone expects a compact domestic format (e.g.
	// "0612345678") but visitors commonly type it with spaces ("06 12 34 56 78").
	request.ContactPhone = strings.Join(strings.Fields(request.ContactPhone), "")
	request.Message = strings.TrimSpace(request.Message)

	if validationErr := validateContactForm(request); validationErr != nil {
		httpx.RespondWithError(w, validationErr, http.StatusBadRequest)
		return
	}

	if err := h.svc.SendContactFormNotification(ctx, request.Name, request.ContactEmail, request.ContactPhone, request.Message); err != nil {
		httpx.RespondWithServiceError(w, logger, ctx, err, "submit contact form")
		return
	}

	logger.InfoContext(ctx, "Handler: Contact form submitted",
		"operation", "submit_contact",
		"status_code", http.StatusAccepted)

	httpx.RespondWithJSON(w, struct {
		Status string `json:"status"`
	}{Status: "sent"}, http.StatusAccepted)
}

func validateContactForm(request contactFormRequest) error {
	if request.Name == "" {
		return errs.NewInvalidValueErr("name is required")
	}
	if len(request.Name) > contactNameMaxLength {
		return errs.NewInvalidValueErr(fmt.Sprintf("name cannot exceed %d characters", contactNameMaxLength))
	}

	if request.Message == "" {
		return errs.NewInvalidValueErr("message is required")
	}
	if len(request.Message) > contactMessageMaxLength {
		return errs.NewInvalidValueErr(fmt.Sprintf("message cannot exceed %d characters", contactMessageMaxLength))
	}

	if request.ContactEmail == "" && request.ContactPhone == "" {
		return errs.NewInvalidValueErr("either contact_email or contact_phone is required")
	}

	if request.ContactEmail != "" {
		if err := validation.ValidateEmail(request.ContactEmail); err != nil {
			return errs.NewInvalidValueErr(fmt.Sprintf("contact_email: %v", err))
		}
	}
	if request.ContactPhone != "" {
		if err := validation.ValidatePhone(request.ContactPhone); err != nil {
			return errs.NewInvalidValueErr(fmt.Sprintf("contact_phone: %v", err))
		}
	}

	return nil
}
