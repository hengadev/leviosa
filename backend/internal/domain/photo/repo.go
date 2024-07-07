package photo

import (
	"context"
	"mime/multipart"
)

type Reader interface{}
type Writer interface {
	AddFile(ctx context.Context, file multipart.File, key string) (string, error)
}

type ReadWriter interface {
	Reader
	Writer
}
