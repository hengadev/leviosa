package mediaService

import (
	"context"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Reader interface {
	FindAllObjects(ctx context.Context, eventID string) ([]types.Object, error)
}

type Writer interface {
	AddDatabaseReplica(ctx context.Context, file multipart.File) (string, error)
}

type ReadWriter interface {
	Reader
	Writer
}

// TODO: The test interface displays the functions that I need to implement
type ReaderTest interface {
	GetEventBanner(ctx context.Context, eventID string) error
	GetEventPhotosByUserID(ctx context.Context, userID, eventID string) ([]types.Object, error)
	GetExerciceVideo(ctx context.Context, name string) error
}

type WriterTest interface {
	AddEventBanner(ctx context.Context, file multipart.File, eventID string) error
	AddUserAvatar(ctx context.Context, file multipart.File, userID string) error
	// TODO: find a better name and signature for the following one
	AddDatabaseReplica(ctx context.Context, file multipart.File) error
	AddExerciceVideo(ctx context.Context, file multipart.File, name string) error
}
