package mediaRepository

import (
	"context"

	rp "github.com/hengadev/leviosa/internal/repository"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (r *Repository) addFile(ctx context.Context, key string) error {
	_, err := r.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(r.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return rp.NewNotCreatedErr(err, "S3 folder")
	}
	return nil
}
