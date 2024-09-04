package mail

import (
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
)

// Function that send an email to user to remind them of an event incoming.
func HandleRemainderEventMail(user *userService.User, eventTime string, daysLeft int) {
	// TODO: Add the call to a certain function to handle using that value
	switch daysLeft {
	case 2:
	case 7:
	}
}
