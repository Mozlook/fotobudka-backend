package storage

import (
	"github.com/Mozlook/fotobudka-backend/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	minio      *minio.Client
	bucketName string
}

func New(cfg config.S3Config) (*Client, error) {
	endpoint := cfg.Endpoint
	accessKeyID := cfg.AccessKeyID
	secretAccessKey := cfg.SecretAccessKey
	useSSL := cfg.UseSSL
	bucketName := cfg.Bucket

	miniClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &Client{minio: miniClient, bucketName: bucketName}, nil
}
