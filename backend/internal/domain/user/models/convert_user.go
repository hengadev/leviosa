package models

func (u User) ToUserResponse() UserResponse {
	return UserResponse{
		Role:      u.Role,
		BirthDate: u.BirthDate.String(),
		LastName:  u.LastName,
		FirstName: u.FirstName,
		Gender:    u.Gender,
		Telephone: u.Telephone,
	}
}

func (u User) ToUserPending() *UserPending {
	return &UserPending{
		EmailHash: u.EmailHash,
		LastName:  u.LastName,
		FirstName: u.FirstName,
		GoogleID:  u.GoogleID,
		AppleID:   u.AppleID,
	}
}
