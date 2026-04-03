package storage

import (
	"github.com/Mozlook/fotobudka-backend/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	internalMinio *minio.Client
	publicMinio   *minio.Client
	bucketName    string
}

func New(cfg config.S3Config) (*Client, error) {
	endpoint := cfg.Endpoint
	publicEndpoint := cfg.EndpointPublic
	if publicEndpoint == "" {
		publicEndpoint = cfg.Endpoint
	}
	accessKeyID := cfg.AccessKeyID
	secretAccessKey := cfg.SecretAccessKey
	useSSL := cfg.UseSSL
	bucketName := cfg.Bucket
	region := cfg.Region
	bucketLookup := minio.BucketLookupAuto
	if cfg.UsePathStyle {
		bucketLookup = minio.BucketLookupPath
	}

	internalMiniClient, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure:       useSSL,
		Region:       region,
		BucketLookup: bucketLookup,
	})
	if err != nil {
		return nil, err
	}
	publicMiniClient, err := minio.New(publicEndpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure:       useSSL,
		Region:       region,
		BucketLookup: bucketLookup,
	})
	if err != nil {
		return nil, err
	}

	return &Client{internalMinio: internalMiniClient, publicMinio: publicMiniClient, bucketName: bucketName}, nil
}
