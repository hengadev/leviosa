package mediaRepository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	BUCKETNAME = "test-bucket-golang-gary"
)

type Repository struct {
	Uploader *manager.Uploader
	Client   *s3.Client
}

func NewRepository(ctx context.Context) (*Repository, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to load default config for the store repository: %w", err)
	}
	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)
	return &Repository{
		Uploader: uploader,
		Client:   client,
	}, nil
}
