package mailService

import (
	"os"
	"path/filepath"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/pkg/errsx"
)

// Function that send an email to user after receiving payment.
func (s *Service) NewVote(user *models.User, eventTime string) errsx.Map {
	var errs errsx.Map
	companyMail, password := getCompanyCredentials()
	templData := struct {
		Username string
		Heure    string
	}{Username: user.FirstName, Heure: eventTime}

	// handle image
	wd, err := os.Getwd()
	if err != nil {
		errs.Set("get working directory", err)
	}
	logoPath := filepath.Join(wd, "internal", "domain", "mail", "assets", "logo.jpg")
	instaPath := filepath.Join(wd, "internal", "domain", "mail", "assets", "instagram.png")

	if err := sendMail(
		companyMail,
		user.Email,
		"[Leviosa] Nouveau votes disponibles",
		"/internal/domain/mail/newRegistry.html",
		password,
		templData,
		map[string]string{
			logoPath: "logo",
			// NOTE: got the link from the instagram logo : https://www.instagram.com/leviosa_care/
			instaPath: "instagram",
		},
	); err != nil {
		errs.Set("send mail", err)
	}
	return errs
}
