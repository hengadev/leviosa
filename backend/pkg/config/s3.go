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

func (c *Config) setS3(env mode.EnvMode) error {
	bucketName := c.viper.GetString("s3.bucketname")
	if bucketName == "" {
		return errors.New("'BUCKETNAME' environment variable not set; please define it to specify S3 bucket name")
	}
	switch env {
	case mode.ModeDev:
		c.s3.BucketName = fmt.Sprintf("staging-%s", bucketName)
	case mode.ModeProd, mode.ModeStaging:
		c.s3.BucketName = fmt.Sprintf("%s-%s", env.String(), bucketName)
	default:
		return fmt.Errorf("mode value can only be 'development', 'production' or 'staging', got : %q", env)
	}
	return nil
}
