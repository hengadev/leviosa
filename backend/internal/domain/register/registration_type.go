package registerService

type RegistrationType string

const (
	Consultation RegistrationType = "consultation"
	Event        RegistrationType = "event"
	AtHome       RegistrationType = "at_home"
)
