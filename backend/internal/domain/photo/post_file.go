package photo

import (
	"context"
	"fmt"
	"mime/multipart"
)

func (s *Service) PostFile(ctx context.Context, file multipart.File, filename, eventID string) (string, error) {
	key := fmt.Sprintf("%s/%s", eventID, filename)
	url, err := s.Repo.AddFile(ctx, file, key)
	if err != nil {
		return "", fmt.Errorf("add file: %w", err)
	}
	return url, nil
}
