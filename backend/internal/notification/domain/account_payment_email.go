package domain

import (
	"fmt"
	"time"
)

// The emails sent outside the booking flow: the login code, the payment
// receipt and failed payment notice, the contact form message and the
// payment dispute alert. Their wording is unchanged from when the SMTP client
// built them inline; the field names are what the templates read.

// Template names of these emails (templates/<name>.html).
const (
	OTPTemplate                 = "otp"
	PaymentTemplate             = "payment"
	PaymentFailedTemplate       = "payment_failed"
	ContactFormTemplate         = "contact_form"
	PaymentDisputeAlertTemplate = "payment_dispute_alert"
)

// OTPEmailData is the data of the login code email.
type OTPEmailData struct {
	OTP     string
	Year    int
	LogoURL string
	// Address and InstagramPath fill the template's footer, which hides
	// them when empty. Nothing sets them yet: before they existed, the
	// template failed to render and no login code email could be sent.
	Address       string
	InstagramPath string
}

// ComposeOTPEmail builds the login code email (English).
func ComposeOTPEmail(req OTPEmailRequest, now time.Time) ComposedEmail {
	return ComposedEmail{
		Subject:  "Your OTP Code",
		Template: OTPTemplate,
		Data: OTPEmailData{
			OTP:     req.OTP,
			Year:    now.Year(),
			LogoURL: req.LogoURL,
		},
	}
}

// PaymentEmailData is the data of the payment receipt.
type PaymentEmailData struct {
	FirstName   string
	LastName    string
	Amount      string
	Product     string
	PaymentDate string
	CompanyName string
	Year        int
	LogoURL     string
}

// PaymentFailedEmailData is the data of the failed payment notice: the
// receipt's, plus the link to try again.
type PaymentFailedEmailData struct {
	PaymentEmailData
	RetryURL string
}

func paymentEmailData(req PaymentNotificationRequest, now time.Time) PaymentEmailData {
	return PaymentEmailData{
		FirstName:   req.ToFirstName,
		LastName:    req.ToLastName,
		Amount:      req.Amount,
		Product:     req.Product,
		PaymentDate: req.PaymentDate,
		CompanyName: req.CompanyName,
		Year:        now.Year(),
		LogoURL:     req.LogoURL,
	}
}

// ComposePaymentNotificationEmail builds the payment receipt (English).
func ComposePaymentNotificationEmail(req PaymentNotificationRequest, now time.Time) ComposedEmail {
	return ComposedEmail{
		Subject:  "Payment Confirmation",
		Template: PaymentTemplate,
		Data:     paymentEmailData(req, now),
	}
}

// ComposePaymentFailedEmail builds the failed payment notice (English), with
// the link to try again.
func ComposePaymentFailedEmail(req PaymentNotificationRequest, now time.Time) ComposedEmail {
	return ComposedEmail{
		Subject:  "Payment Failed",
		Template: PaymentFailedTemplate,
		Data:     PaymentFailedEmailData{PaymentEmailData: paymentEmailData(req, now), RetryURL: req.RetryURL},
	}
}

// ContactFormEmailData is the data of the contact form message.
type ContactFormEmailData struct {
	Name         string
	ContactEmail string
	ContactPhone string
	Message      string
	SubmittedAt  string
	CompanyName  string
	Year         int
	LogoURL      string
}

// ComposeContactFormEmail builds the contact form message to the practice
// (French). Replies go to the visitor when they left an email.
func ComposeContactFormEmail(req ContactFormRequest, now time.Time) ComposedEmail {
	return ComposedEmail{
		Subject:  fmt.Sprintf("Nouveau message de contact — %s", req.Name),
		Template: ContactFormTemplate,
		Data: ContactFormEmailData{
			Name:         req.Name,
			ContactEmail: req.ContactEmail,
			ContactPhone: req.ContactPhone,
			Message:      req.Message,
			SubmittedAt:  req.SubmittedAt,
			CompanyName:  req.CompanyName,
			Year:         now.Year(),
			LogoURL:      req.LogoURL,
		},
		ReplyTo: req.ContactEmail,
	}
}

// PaymentDisputeAlertEmailData is the data of the payment dispute alert.
type PaymentDisputeAlertEmailData struct {
	BookingID     string
	ClientName    string
	ProductName   string
	Amount        string
	DisputeReason string
	Year          int
}

// ComposePaymentDisputeAlertEmail builds the payment dispute alert to the
// practice (English).
func ComposePaymentDisputeAlertEmail(req PaymentDisputeAlertRequest, now time.Time) ComposedEmail {
	return ComposedEmail{
		Subject:  fmt.Sprintf("Payment dispute opened — booking %s", req.BookingID),
		Template: PaymentDisputeAlertTemplate,
		Data: PaymentDisputeAlertEmailData{
			BookingID:     req.BookingID,
			ClientName:    req.ClientName,
			ProductName:   req.ProductName,
			Amount:        req.Amount,
			DisputeReason: req.DisputeReason,
			Year:          now.Year(),
		},
	}
}
