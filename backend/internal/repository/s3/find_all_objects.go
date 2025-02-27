package mediaRepository

import (
	"context"
	"fmt"

	rp "github.com/GaryHY/leviosa/internal/repository"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (r *Repository) FindAllObjects(ctx context.Context, eventID string) ([]types.Object, error) {
	res := make([]types.Object, 0)
	localisation := fmt.Sprintf("%s/%s", r.BucketName, eventID)
	output, err := r.Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(localisation),
	})
	if err != nil {
		return nil, rp.NewNotFoundErr(err, "photos")
	}
	for _, object := range output.Contents {
		_ = object
		// res = append(res, object)
	}
	return res, nil
}
