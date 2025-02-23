package mailService

import "github.com/GaryHY/leviosa/pkg/errsx"

func (s *Service) HandlePasswordForgotten() errsx.Map {
	var errs errsx.Map
	// send an email to the user and when redirected to that link, give the user an opportunity to remake the password.
	return errs
}
