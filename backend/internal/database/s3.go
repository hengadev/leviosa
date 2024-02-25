package sqlite

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
	"mime/multipart"
)

const (
	BUCKETNAME = "test-bucket-golang-gary"
)

type PhotoStore struct {
	Uploader *manager.Uploader
}

func NewPhotoStore() *PhotoStore {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("error: %v", err)
	}
	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)
	return &PhotoStore{Uploader: uploader}
}

func (p *PhotoStore) PostFile(file multipart.File, filename, event_id string) {
	key := fmt.Sprintf("%s/%s", event_id, filename)
	result, err := p.Uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(BUCKETNAME),
		Key:    aws.String(key),
		Body:   file,
		ACL:    "public-read",
	})
	if err != nil {
		log.Fatal("cannot send the file to the bucket", err)
	}
	_ = result
	// fmt.Println("The result of the upload is found at : ", result.Location)
}
