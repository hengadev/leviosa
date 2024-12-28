package mailService

import (
	"context"
	"os"
	"path/filepath"

	otpService "github.com/GaryHY/event-reservation-app/internal/domain/otp"
	"github.com/GaryHY/event-reservation-app/pkg/errsx"
)

// TODO: make the right email template for that mail domain service
func (s *Service) SendOTP(ctx context.Context, email, firstname string, otp *otpService.OTP) errsx.Map {
	var errs errsx.Map
	companyMail, password := getCompanyCredentials()

	wd, err := os.Getwd()
	if err != nil {
		errs.Set("get working directory: %s", err)
	}

	logoPath := filepath.Join(wd, "internal", "domain", "mail", "assets", "logo.jpg")
	instaPath := filepath.Join(wd, "internal", "domain", "mail", "assets", "instagram.png")

	// data used in the email
	templData := struct {
		Firstname string
		Value     string
	}{
		Firstname: firstname,
		Value:     otp.Code,
	}
	if err := sendMail(
		companyMail,
		email,
		"[Leviosa] Confirmation d'addresse email",
		"/internal/domain/mail/templates/otp.html",
		password,
		templData,
		map[string]string{
			logoPath:  "logo",
			instaPath: "instagram",
		},
	); err != nil {
		errs.Set("send email:", err)
	}
	return errs
}
