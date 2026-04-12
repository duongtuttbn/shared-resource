package s3

import "io"

type Config struct {
	AccessKeyID     string `json:"access_key_id" mapstructure:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key" mapstructure:"secret_access_key"`
	Region          string `json:"region" mapstructure:"region"`
	Bucket          string `json:"bucket" mapstructure:"bucket"`
	CdnURL          string `json:"cdn_url" mapstructure:"cdn_url"`
	Endpoint        string `json:"endpoint" mapstructure:"endpoint"`
}

type readerWithLength struct {
	io.Reader
	length int
}

func (r *readerWithLength) Len() int {
	return r.length
}
