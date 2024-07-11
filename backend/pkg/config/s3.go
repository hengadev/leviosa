package config

type s3Creds struct {
}

func GetS3() *s3Creds {
	return &s3Creds{}
}

func setS3() error {
	return nil
}
