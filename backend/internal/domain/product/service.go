package productService

type Service struct {
	repo ReadWriter
}

func New(repo ReadWriter) *Service {
	return &Service{repo}
}
