package userService

type Role int8

const (
	UNKNOWN Role = iota
	BASIC
	GUEST
	ADMINISTRATOR
)

var roles = [4]string{
	"unknown",
	"basic",
	"guest",
	"admin",
}

func (r Role) String() string {
	return roles[r]
}

func ConvertToRole(role string) Role {
	switch role {
	case "admin":
		return ADMINISTRATOR
	case "guest":
		return GUEST
	case "basic":
		return BASIC
	default:
		return UNKNOWN
	}

	// case "":
	// 	return UNKNOWN
	// default:
	// 	return BASIC

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
