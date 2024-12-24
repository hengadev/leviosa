package mailService

import (
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
)

// Function that send an email to user to remind them of an event incoming.
func (s *Service) HandleRemainderEventMail(user *models.User, eventTime string, daysLeft int) {
	// TODO: Add the call to a certain function to handle using that value
	switch daysLeft {
	case 2:
	case 7:
	}
}
