package types

// The link for how to implement enums in golang : https://www.sohamkamani.com/golang/enums/
type Role int8

func ConvertToRole(str string) Role {
	switch str {
	case "admin":
		return ADMIN
	case "helper":
		return HELPER
	case "basic":
		return BASIC
	default:
		return UNKNOWN
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

func (r Role) String() string {
	roles := []string{
		"unknown",
		"user",
		"guest",
		"administrator",
	}
	return roles[r]
}

const (
	UNKNOWN Role = iota // the default value too
	BASIC
	HELPER
	ADMIN
)
