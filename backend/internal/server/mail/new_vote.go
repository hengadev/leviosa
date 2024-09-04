package mail

// Function that send an email to user after receiving payment.
func NewVote(user *userService.User, eventTime string) {
	companyMail, password := getCompanyCredentials()
	templData := struct {
		Username string
		Heure    string
	}{Username: user.FirstName, Heure: eventTime}
	sendMail(companyMail, user.Email, "Bienvenue parmi nous !", "/internal/mail/newRegistry.html", password, templData)
}
