package otpService

type Service struct {
	Repo ReadWriter
}

func New(repo ReadWriter) *Service {
	return &Service{
		Repo: repo,
	}
}
