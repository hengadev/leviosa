package media

import (
	"context"
	"errors"
	"io"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

func (r *repository) PutObject(ctx context.Context, key string, body io.ReadSeeker, size int64, contentType string) error {
	_, err := r.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r.BucketName),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return errs.NewExternalStorageErr(err, "S3 PutObject for media", key)
	}
	return nil
}

func (r *repository) GetObject(ctx context.Context, key, rangeHeader string) (*domain.MediaObject, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(r.BucketName),
		Key:    aws.String(key),
	}
	if rangeHeader != "" {
		input.Range = aws.String(rangeHeader)
	}
	out, err := r.Client.GetObject(ctx, input)
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return nil, errs.NewRepositoryNotFoundErr(err, "media "+key)
		}
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "InvalidRange" {
			return nil, domain.ErrMediaRangeNotSatisfiable
		}
		return nil, errs.NewExternalStorageErr(err, "S3 GetObject for media", key)
	}
	return &domain.MediaObject{
		Body:          out.Body,
		ContentType:   aws.ToString(out.ContentType),
		ContentLength: aws.ToInt64(out.ContentLength),
		ContentRange:  aws.ToString(out.ContentRange),
		ETag:          aws.ToString(out.ETag),
		LastModified:  out.LastModified,
	}, nil
}

func (r *repository) DeleteObject(ctx context.Context, key string) error {
	_, err := r.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return errs.NewExternalStorageErr(err, "S3 DeleteObject for media", key)
	}
	return nil
}
