package mediaRepository

import (
	"context"
	"mime/multipart"
)

func (r *Repository) AddDatabaseReplica(ctx context.Context, file multipart.File) (string, error) {
	location, err := r.addMultipartFile(ctx, file, "")
	if err != nil {
		return "", err
	}
	return location, nil
}
