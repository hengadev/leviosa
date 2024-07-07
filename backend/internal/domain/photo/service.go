package photo

type Service struct {
	Repo ReadWriter
}

func NewService(repo ReadWriter) *Service {
	return &Service{repo}
}
