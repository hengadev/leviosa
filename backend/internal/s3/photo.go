package photo

import (
	"context"
	"fmt"
	"mime/multipart"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	BUCKETNAME = "test-bucket-golang-gary"
)

type PhotoRepository struct {
	Uploader *manager.Uploader
	Client   *s3.Client
}

func NewPhotoRepository(ctx context.Context) (*PhotoRepository, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to load default config for the store repository: %w", err)
	}
	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)
	return &PhotoRepository{
		Uploader: uploader,
		Client:   client,
	}, nil
}

func (p *PhotoRepository) AddFile(ctx context.Context, file multipart.File, key string) (string, error) {
	result, err := p.Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(BUCKETNAME),
		Key:    aws.String(key),
		Body:   file,
		ACL:    "public-read",
	})
	if err != nil {
		return "", rp.NewRessourceCreationErr(err)
	}
	return result.Location, nil
}

func (p *PhotoRepository) FindAllObjects(ctx context.Context, eventID string) ([]types.Object, error) {
	res := make([]types.Object, 0)
	localisation := fmt.Sprintf("%s/%s", BUCKETNAME, eventID)
	output, err := p.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(localisation),
	})
	if err != nil {
		return nil, rp.NewNotFoundError(err)
	}
	for _, object := range output.Contents {
		res = append(res, object)
	}
	return res, nil
}
