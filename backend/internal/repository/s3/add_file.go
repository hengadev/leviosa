package mediaRepository

import (
	"context"
	"mime/multipart"

	rp "github.com/GaryHY/leviosa/internal/repository"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (r *Repository) AddFile(ctx context.Context, file multipart.File, key string) (string, error) {
	result, err := r.Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(BUCKETNAME),
		Key:    aws.String(key),
		Body:   file,
		// TODO: I do not think that the content here shouold be public read
		ACL: "public-read",
	})
	if err != nil {
		return "", rp.NewNotCreatedErr(err, "photo")
	}
	return result.Location, nil
}
