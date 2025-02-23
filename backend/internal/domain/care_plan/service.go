package carePlanService

type Service struct {
	Repo ReadWriter
}

func New(repo ReadWriter) *Service {
	return &Service{repo}
}
