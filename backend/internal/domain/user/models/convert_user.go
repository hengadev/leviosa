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
		LastName:  u.LastName,
		FirstName: u.FirstName,
	}
}
