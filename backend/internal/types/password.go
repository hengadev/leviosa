package types

type Password string

func NewPassword(password string) (Password, error) {
	// TODO: make some validation for the password too.
	return Password(password), nil
}

func (p Password) String() string {
	return string(p)
}
