package sqlite

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// TODO: Put that in an .env file ?
const (
	BUCKETNAME = "test-bucket-golang-gary"
)

type PhotoStore struct {
	Uploader *manager.Uploader
	Client   *s3.Client
}

func NewPhotoStore() *PhotoStore {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("error: %v", err)
	}
	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)
	return &PhotoStore{
		Uploader: uploader,
		Client:   client,
	}
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

// TODO:
// https://dev.to/aws-builders/get-objects-from-aws-s3-bucket-with-golang-2mne
func (p *PhotoStore) GetAllobjects(event_id string) []types.Object {
	res := make([]types.Object, 0)
	localisation := fmt.Sprintf("%s/%s", BUCKETNAME, event_id)
	output, err := p.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(localisation),
	})
	if err != nil {
		log.Fatal("error getting the result from the bucket - ", err)
	}
	for _, object := range output.Contents {
		res = append(res, object)
	}
	// return make([]string, 0)
	return res
}
