package userService

import (
	"context"
	"fmt"
)

func (s *Service) CreateOAuthAccount(
	ctx context.Context,
	userCandidate OAuthUser,
) (*User, error) {
	fmt.Println("just before the ToUser conversion")
	user, err := userCandidate.ToUser()
	if err != nil {
		return nil, fmt.Errorf("convert OAuthUser to User")
	}
	fmt.Println("convert googleUser to User")
	// NOTE: We use the same function as the email-password one but we need to make the function better to do account linking.
	lastInsertID, err := s.repo.AddAccount(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("add oauth account: %w", err)
	}
	fmt.Println("user inserted in database")
	user.ID = int(lastInsertID)
	user.Role = BASIC.String()
	user.Create()
	user.Login()
	fmt.Printf("the user is after the conversion from the OAuth: %#+v\n", user)
	return user, nil
}
