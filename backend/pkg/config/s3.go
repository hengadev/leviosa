package config

import (
	"errors"
	"fmt"

	mode "github.com/GaryHY/leviosa/pkg/flags"
)

type s3Creds struct {
	BucketName string `json:"bucketname"`
}

func (c *Config) GetS3() *s3Creds {
	return c.s3
}

func (c *Config) GetBucketName() string {
	return c.s3.BucketName
}

func (c *Config) setS3(env mode.EnvMode) error {
	bucketName := c.viper.GetString("s3.bucketname")
	if bucketName == "" {
		return errors.New("'BUCKETNAME' environment variable not set; please define it to specify S3 bucket name")
	}
	c.s3.BucketName = fmt.Sprintf("%s_%s", env.String(), bucketName)
	return nil
}
