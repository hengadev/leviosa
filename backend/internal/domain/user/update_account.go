package userService

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/pkg/errsx"
	"github.com/google/uuid"
)

func (s *Service) UpdateAccount(ctx context.Context, userCandidate *User, userID string) error {
	if err := uuid.Validate(userID); err != nil {
		return domain.NewInvalidValueErr(fmt.Sprintf("invalid user ID: %s", err.Error()))
	}

	if pbms := userCandidate.updateValid(); len(pbms) > 0 {
		return domain.NewInvalidValueErr(fmt.Sprintf("%s", pbms.Error()))
	}

	// TODO: encrypt the user data here

	err := s.repo.ModifyAccount(
		ctx,
		userCandidate,
		map[string]any{"id": userID},
		prohibitedFields...,
	)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrInternal):
			fallthrough
		case errors.Is(err, rp.ErrContext):
			return err
		case errors.Is(err, rp.ErrDatabase):
			return domain.NewQueryFailedErr(err)
		case errors.Is(err, rp.ErrNotUpdated):
			return domain.NewNotUpdatedErr(err)
		default:
			return domain.NewUnexpectTypeErr(err)
		}
	}
	return nil
}

func (u User) updateValid() errsx.Map {
	var pbms errsx.Map
	vf := reflect.VisibleFields(reflect.TypeOf(u))
	for _, f := range vf {
		switch f.Name {
		case "Telephone":
			if len(u.Telephone) < 10 {
				pbms.Set("telephone too short", fmt.Errorf(""))
			}
			if !strings.HasPrefix(u.Telephone, "0") {
				pbms.Set("telephone does not start with 0", fmt.Errorf(""))
			}
		case "Birthday":
			parsedDate, err := time.Parse(BirthdayLayout, u.BirthDate)
			nonValidDate, _ := time.Parse(BirthdayLayout, "01-01-01")
			if err != nil && parsedDate != nonValidDate {
				pbms.Set("birthday", err)
			}
		case "Lastname":
			if u.LastName == "" {
				pbms.Set("lastname", errors.New("empty lastname"))
			}
		case "Firstname":
			if u.FirstName == "" {
				pbms.Set("firstname", errors.New("empty lastname"))
			}
		default:
			continue
		}
	}
	return pbms
}
