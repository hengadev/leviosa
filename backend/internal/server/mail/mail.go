package mail

import (
	"bytes"
	"html/template"
	"os"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"

	"gopkg.in/gomail.v2"
)

// TODO: use the right user type for this, you just have to use the struct tag better

// Function that send an email to all users to specify that a new event has been created.
func HandleNewEventMail(usersList []*user.User, eventTime string) {
	companyMail, password := getCompanyCredentials()
	for _, user := range usersList {
		go func(firstname string) {
			templData := struct {
				Username string
				Heure    string
			}{Username: user.FirstName, Heure: eventTime}
			sendMail(companyMail, user.Email, "Un nouvel evenement pourrait vous interesser", "/internal/mail/templates/newEvent.html", password, templData)
		}(user.FirstName)
	}
}

// Function that send an email to user after receiving payment.
func SendNewVoteMail(user *user.User, eventTime string) {
	companyMail, password := getCompanyCredentials()
	templData := struct {
		Username string
		Heure    string
	}{Username: user.FirstName, Heure: eventTime}
	sendMail(companyMail, user.Email, "Bienvenue parmi nous !", "/internal/mail/newRegistry.html", password, templData)
}

func HandleNewPaymentMail(user *user.User, eventTime string) {
}

// Function that send an email to user to remind them of an event incoming.
func HandleRemainderEventMail(user *user.User, eventTime string, daysLeft int) {
	// TODO: Add the call to a certain function to handle using that value
	switch daysLeft {
	case 2:
	case 7:
	}
}

func HandleRemainderPaymentMail(user *user.User, eventTime string) {
	companyMail, password := getCompanyCredentials()
	templData := struct {
		Username string
	}{Username: user.FirstName}
	sendMail(companyMail, user.Email, "Bienvenue parmi nous !", "/internal/mail/welcomeUser.html", password, templData)
}

func SendWelcomeUserMail(user *user.User) {
}

func sendMail(from, to, subject, templateFilename, password string, data any) {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	// m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", subject)

	// get the working directory to get the file to parse (the one with the template)
	wd, _ := os.Getwd()
	t, _ := template.ParseFiles(wd + templateFilename)

	var tpl bytes.Buffer
	t.Execute(&tpl, data)
	m.SetBody("text/html", tpl.String())
	// m.Attach("/home/Alex/lolcat.jpg")

	d := gomail.NewDialer("smtp.gmail.com", 587, from, password)

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

func getCompanyCredentials() (string, string) {
	return os.Getenv("GMAIL_EMAIL"), os.Getenv("GMAIL_PASSWORD")
}

// TODO: A function to get back a forgottend password.
func HandlePasswordForgotten() {
	// send an email to the user and when redirected to that link, give the user an opportunity to remake the password.
}

// FIX:
// 0. Learn how to send a simple mail.
// TODO:
// all the rest is handle in the store
// 1. parse the events tables to get a list of event that I need to send an email to for each rappelLimit (using cron jobs)
// 2. for each event get a list of user that voted for them
// 3. send an email for each user with the right template depending on the rappelLimit.
