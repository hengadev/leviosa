package mediaRepository

import (
	"context"
	"mime/multipart"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (p *PhotoRepository) AddFile(ctx context.Context, file multipart.File, key string) (string, error) {
	result, err := p.Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(BUCKETNAME),
		Key:    aws.String(key),
		Body:   file,
		ACL:    "public-read",
	})
	if err != nil {
		return "", rp.NewNotCreatedErr(err, "photo")
	}
	return result.Location, nil
}
