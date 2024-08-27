package userService

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (s *Service) UpdateAccount(ctx context.Context, userCandidate *User, userID int) error {
	// TODO: handle that valid function
	if pbms := userCandidate.SmallValid(ctx); len(pbms) > 0 {
		return serverutil.FormatError(pbms, "user")
	}
	err := s.repo.ModifyAccount(
		ctx,
		userCandidate,
		map[string]any{"id": userID},
		"ID",
		"Email",
		"Password",
		"CreatedAt",
		"LoggedInAt",
		"Role",
	)
	if err != nil {
		return fmt.Errorf("add account: %w", err)
	}
	return nil
}

func (u User) SmallValid(ctx context.Context) map[string]string {
	var pbms = make(map[string]string)
	vf := reflect.VisibleFields(reflect.TypeOf(u))
	for _, f := range vf {
		switch f.Name {
		case "Telephone":
			// do the validation using the rule that follows :
			// if len(u.Telephone) < 10 && strings.HasPrefix(u.Telephone) {
			// 	pbms["telephone"] = ""
			// }
		case "Birthday":
			parsedDate, err := time.Parse(BirthdayLayout, u.BirthDate)
			nonValidDate, _ := time.Parse(BirthdayLayout, "01-01-01")
			if err != nil && parsedDate != nonValidDate {
				pbms["birthday"] = err.Error()
			}
		default:
			continue
		}
	}
	return pbms
}
