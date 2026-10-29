package domain

import "time"

// OTP Email
type OTPEmailRequest struct {
	ToEmail   string `json:"to_email"`
	OTP       string `json:"otp"`
	FromEmail string `json:"from_email"`
	LogoURL   string `json:"logo_url"`
}

// Welcome Email
type WelcomeEmailRequest struct {
	ToEmail     string `json:"to_email"`
	ToFirstName string `json:"to_first_name"`
	ToLastName  string `json:"to_last_name"`
	FromEmail   string `json:"from_email"`
	CompanyName string `json:"company_name"`
	LogoURL     string `json:"logo_url"`
}

// Verify Email
type VerifyEmailRequest struct {
	ToEmail     string `json:"to_email"`
	ToFirstName string `json:"to_first_name"`
	ToLastName  string `json:"to_last_name"`
	FromEmail   string `json:"from_email"`
	CompanyName string `json:"company_name"`
	LogoURL     string `json:"logo_url"`
}

// Event Notification
type EventNotificationRequest struct {
	ToEmail     string `json:"to_email"`
	ToFirstName string `json:"to_first_name"`
	ToLastName  string `json:"to_last_name"`
	Event       string `json:"event"`
	Details     string `json:"details"`
	FromEmail   string `json:"from_email"`
	CompanyName string `json:"company_name"`
	LogoURL     string `json:"logo_url"`
}

// Payment Notification
type PaymentNotificationRequest struct {
	ToEmail     string `json:"to_email"`
	ToFirstName string `json:"to_first_name"`
	ToLastName  string `json:"to_last_name"`
	Amount      string `json:"amount"`
	Product     string `json:"product"`
	PaymentDate string `json:"payment_date"`
	FromEmail   string `json:"from_email"`
	CompanyName string `json:"company_name"`
	LogoURL     string `json:"logo_url"`
	RetryURL    string `json:"retry_url,omitempty"` // non-empty for payment_failed emails
}

// BookingReminderRequest is the client reminder sent 24 hours before a
// confirmed booking, written in the locale the client booked in (composed by
// ComposeBookingReminder). Like every booking email it carries no health
// data (ADR-0019).
type BookingReminderRequest struct {
	ToEmail     string
	ToFirstName string
	ToLastName  string
	Locale      Locale

	BookingID   string
	ProductName string
	PartnerName string
	// Address is the practice address.
	Address string

	StartTime time.Time
	EndTime   time.Time

	// ManageURL opens the booking token page. Empty when the booking has no
	// token.
	ManageURL string
}

// Contact Form Notification — sent to staff/company inbox when a visitor
// submits the public contact form. Reuses the existing SMTP notification
// channel rather than a new inbound-message pipeline.
type ContactFormRequest struct {
	ToEmail      string `json:"to_email"`   // staff/company inbox (recipient)
	FromEmail    string `json:"from_email"` // company email (sender)
	Name         string `json:"name"`
	ContactEmail string `json:"contact_email,omitempty"` // visitor's email, used as Reply-To when present
	ContactPhone string `json:"contact_phone,omitempty"`
	Message      string `json:"message"`
	CompanyName  string `json:"company_name"`
	LogoURL      string `json:"logo_url"`
	SubmittedAt  string `json:"submitted_at"`
}

// Payment Dispute Alert Email (admin-facing)
type PaymentDisputeAlertRequest struct {
	ToEmail       string `json:"to_email"`
	BookingID     string `json:"booking_id"`
	ClientName    string `json:"client_name"`
	ProductName   string `json:"product_name"`
	Amount        string `json:"amount"`
	DisputeReason string `json:"dispute_reason"`
	FromEmail     string `json:"from_email"`
}

// EmailRequest is the generic email structure used by SMTP client
type EmailRequest struct {
	To         string
	ReplyTo    string
	Subject    string
	Template   string
	Data       interface{}
	CarbonCopy map[string]string // email -> name
	Images     map[string]string // path -> rename
	// Attachments are added as attachment parts (multipart/mixed).
	Attachments []EmailAttachment
}

// EmailAttachment is a file attached to an email.
type EmailAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
}
