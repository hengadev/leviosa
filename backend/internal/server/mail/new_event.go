package mail

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
)

// Send an email to all users notifying new event creation.
func NewEvent(users []*userService.User, eventTime string) []error {
	companyMail, password := getCompanyCredentials()
	var wg sync.WaitGroup
	var errList []error
	var errMutex sync.Mutex

	// handle image
	wd, err := os.Getwd()
	if err != nil {
		errList = append(errList, fmt.Errorf("get working directory: %s", err))
	}

	logoPath := filepath.Join(wd, "internal", "mail", "assets", "logo.jpg")
	instaPath := filepath.Join(wd, "internal", "mail", "assets", "instagram.png")

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
				Firstname: user.Firstname,
				Heure:     eventTime,
			}
			for _, email := range emails {
				if err := sendMail(
					companyMail,
					email,
					"[Leviosa] Nouvel Évènement disponible",
					"/internal/mail/templates/newEvent.html",
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
					errList = append(errList, fmt.Errorf("send mail: %s", err))
					errMutex.Unlock()
				}
			}
		}()
	}
	fmt.Println("waiting for emails to be sent to users")
	wg.Wait()
	if len(errList) > 0 {
		return errList
	}
	return nil
}

// Take an array of type with the name of that type and return the plural is the array has a length > 1.
func ToPlural[T any](arr []*T, name string) string {
	if len(arr) == 1 {
		return name
	}
	return name + "s"
}
