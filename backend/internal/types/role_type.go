package types

type Role string

func ConvertToRole(str string) Role {
	switch str {
	case "admin":
		return ADMIN
	case "helper":
		return HELPER
	case "basic":
		return BASIC
	default:
		return BASIC
	}
}

const (
	ADMIN  = Role("admin")
	BASIC  = Role("basic")
	HELPER = Role("helper")
)
