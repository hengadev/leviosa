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

// Function qui retourne si un role est superieur (ou egal a un autre role).
func (r Role) IsSuperior(role Role) bool {
	switch r {
	case ADMIN:
		return role == ADMIN || role == HELPER || role == BASIC
	case HELPER:
		return role == HELPER
		// NOTE: Les helpers sont des intervenants qui ne participent pas aux evenements donc, ils ne peuvent pas voter etc... Mais dans le cas  ou je veux changer
		// return role == HELPER || role == BASIC
	case BASIC:
		return role == BASIC
	default:
		return false
	}
}

const (
	ADMIN  = Role("admin")
	BASIC  = Role("basic")
	HELPER = Role("helper")
)
