package domain

import (
	"fmt"
	"time"
)

// The emails sent outside the booking flow: the login code, the payment
// receipt (no longer sent) and failed payment notice, the contact form
// message and the payment dispute alert. All French; the field names are what
// the templates read.

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
	OTP string
	// ValidForMinutes is stated in the email; 0 says "quelques minutes".
	ValidForMinutes int
	Year            int
}

// ComposeOTPEmail builds the login code email (French), sent for a password
// reset, an email check and a guest claim: its wording fits all three. It
// states how long the code stays valid (the duration set in the admin
// settings).
func ComposeOTPEmail(req OTPEmailRequest, now time.Time) ComposedEmail {
	return ComposedEmail{
		Subject:  "Votre code de vérification Leviosa",
		Template: OTPTemplate,
		Data: OTPEmailData{
			OTP:             req.OTP,
			ValidForMinutes: req.ValidForMinutes,
			Year:            now.Year(),
		},
	}
}

// PaymentEmailData is the data of the payment receipt, which is no longer
// sent (the booking confirmation tells the client their payment went
// through).
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

// ComposePaymentFailedEmail builds the failed payment notice (French), with
// the link to the client's bookings.
func ComposePaymentFailedEmail(req PaymentNotificationRequest, now time.Time) ComposedEmail {
	return ComposedEmail{
		Subject:  "Échec de votre paiement",
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
// practice (French).
func ComposePaymentDisputeAlertEmail(req PaymentDisputeAlertRequest, now time.Time) ComposedEmail {
	return ComposedEmail{
		Subject:  fmt.Sprintf("Litige de paiement — réservation %s", req.BookingID),
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
