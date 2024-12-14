package mail

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
)

// Function that send an email to user after receiving payment.
func NewVote(user *userService.User, eventTime string) {
	companyMail, password := getCompanyCredentials()
	templData := struct {
		Username string
		Heure    string
	}{Username: user.FirstName, Heure: eventTime}
	var errList []error

	// handle image
	wd, err := os.Getwd()
	if err != nil {
		errList = append(errList, fmt.Errorf("get working directory: %s", err))
	}
	logoPath := filepath.Join(wd, "internal", "mail", "assets", "logo.jpg")
	instaPath := filepath.Join(wd, "internal", "mail", "assets", "instagram.png")

	sendMail(
		companyMail,
		user.Email,
		"[Leviosa] Nouveau votes disponibles",
		"/internal/mail/newRegistry.html",
		password,
		templData,
		map[string]string{
			logoPath: "logo",
			// NOTE: got the link from the instagram logo : https://www.instagram.com/leviosa_care/
			instaPath: "instagram",
		},
	)
}
