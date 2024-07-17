package config

import "context"

type s3Creds struct {
	Bucketname string `json:"bucketname"`
}

func (c *Config) GetS3() *s3Creds {
	return c.s3
}

func (c *Config) setS3(context.Context) error {
	c.s3.Bucketname = c.viper.GetString("s3.bucketname")
	return nil
}
