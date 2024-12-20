package mediaRepository

import (
	"context"
	"fmt"

	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (p *PhotoRepository) FindAllObjects(ctx context.Context, eventID string) ([]types.Object, error) {
	res := make([]types.Object, 0)
	localisation := fmt.Sprintf("%s/%s", BUCKETNAME, eventID)
	output, err := p.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(localisation),
	})
	if err != nil {
		return nil, rp.NewNotFoundError(err, "photos")
	}
	for _, object := range output.Contents {
		_ = object
		// res = append(res, object)
	}
	return res, nil
}
