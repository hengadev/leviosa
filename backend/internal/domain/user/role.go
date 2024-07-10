package user

type Role int8

const (
	UNKNOWN       Role = iota
	BASIC         Role = iota
	GUEST         Role = iota
	ADMINISTRATOR Role = iota
)

func (r Role) String() string {
	roles := []string{
		"unknown",
		"basic",
		"guest",
		"admin",
	}
	return roles[r]
}

func ConvertToRole(role string) Role {
	switch role {
	case "admin":
		return ADMINISTRATOR
	case "helper":
		return GUEST
	case "basic":
		return BASIC
	default:
		return UNKNOWN
	}
}

// Function qui retourne si un role est superieur (ou egal a un autre role).
func (r Role) IsSuperior(role Role) bool {
	switch r {
	case ADMINISTRATOR:
		return role == ADMINISTRATOR || role == GUEST || role == BASIC
	case GUEST:
		return role == GUEST
	case BASIC:
		return role == BASIC
	default:
		return false
	}
}
