package mediaService

type Service struct {
	Repo ReadWriter
}

func New(repo ReadWriter) *Service {
	return &Service{repo}
}

// TODO: the thing that I am going to handle with S3
// - crud user picture (for the user that do not use oauth)
// - crud event banner
// - crud events photos
// - crud videos for the exercices that a client can do
// - crud offers photos
// - crud offers videos ?
