package mediaRepository

import (
	"context"

	rp "github.com/GaryHY/leviosa/internal/repository"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// TODO: that thing needs to be gone
func (r *Repository) CreateFolder(ctx context.Context, foldername string) error {
	// Put an empty object with folderName as the key
	_, err := r.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(BUCKETNAME),
		Key:    aws.String(foldername),
	})
	if err != nil {
		return rp.NewNotCreatedErr(err, "S3 folder")
	}
	return nil
}
