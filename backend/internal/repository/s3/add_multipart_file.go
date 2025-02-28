package mediaRepository

import (
	"context"
	"mime/multipart"

	rp "github.com/hengadev/leviosa/internal/repository"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (r *Repository) addMultipartFile(ctx context.Context, file multipart.File, key string) (string, error) {
	result, err := r.Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.BucketName),
		Key:    aws.String(key),
		Body:   file,
		// TODO: set the access control list better to help with speed and security
		ACL: "public-read",
	})
	if err != nil {
		return "", rp.NewNotCreatedErr(err, "S3 bucket object")
	}
	return result.Location, nil
}
