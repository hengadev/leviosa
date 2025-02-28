package mailService

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/hengadev/leviosa/internal/domain/user/models"
	"github.com/hengadev/leviosa/pkg/errsx"
)

// Send an email to all users notifying new event creation.
func (s *Service) NewEvent(users []*models.User, eventTime string) errsx.Map {
	var errs errsx.Map
	companyMail, password := getCompanyCredentials()
	var wg sync.WaitGroup
	var errMutex sync.Mutex

	// handle image
	wd, err := os.Getwd()
	if err != nil {
		errs.Set("get working directory", err)
	}

	logoPath := filepath.Join(wd, "internal", "domain", "mail", "assets", "logo.jpg")
	instaPath := filepath.Join(wd, "internal", "domain", "mail", "assets", "instagram.png")

	userSyntax := ToPlural(users, "user")
	fmt.Printf("sending email to %d %s\n", len(users), userSyntax)

	for _, user := range users {
		wg.Add(1)
		go func() {
			// this is just to test of both clients
			emails := []string{user.Email, "henry.gary@hotmail.com"}
			// here I just test with oulook since it does not work
			// emails := []string{"henry.gary@hotmail.com"}
			defer func() {
				for _, email := range emails {
					fmt.Printf("Finish sending email to %s\n", email)
				}
				wg.Done()
			}()
			templData := struct {
				Firstname string
				Heure     string
			}{
				Firstname: user.FirstName,
				Heure:     eventTime,
			}
			for _, email := range emails {
				if err := sendMail(
					companyMail,
					email,
					"[Leviosa] Nouvel Évènement disponible",
					"/internal/domain/mail/templates/newEvent.html",
					password,
					templData,
					map[string]string{
						logoPath: "logo",
						// NOTE: got the link from the instagram logo : https://www.instagram.com/leviosa_care/
						instaPath: "instagram",
					},
				); err != nil {
					fmt.Printf("error occured sending email to user %s: %s\n", user.Email, err)
					errMutex.Lock()
					errs.Set("send mail", err)
					errMutex.Unlock()
				}
			}
		}()
	}
	fmt.Println("waiting for emails to be sent to users")
	wg.Wait()
	return errs
}

// Take an array of type with the name of that type and return the plural is the array has a length > 1.
func ToPlural[T any](arr []*T, name string) string {
	if len(arr) == 1 {
		return name
	}
	return name + "s"
}
