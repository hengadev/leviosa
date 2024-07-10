package photo

import (
	"context"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Reader interface {
	FindAllObjects(ctx context.Context, eventID string) ([]types.Object, error)
}
type Writer interface {
	AddFile(ctx context.Context, file multipart.File, key string) (string, error)
}

type ReadWriter interface {
	Reader
	Writer
}
