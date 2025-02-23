package models

type UserResponse struct {
	Role      string `json:"role,omitempty"`
	BirthDate string `json:"birthdate,omitempty"`
	LastName  string `json:"lastname,omitempty"`
	FirstName string `json:"firstname,omitempty"`
	Gender    string `json:"gender,omitempty"`
	Telephone string `json:"telephone,omitempty"`
}
