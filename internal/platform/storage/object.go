package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/minio/minio-go/v7"
)

var ErrObjectNotFound = errors.New("object with that key doesn't exist")

func (c *Client) StatObject(ctx context.Context, objectKey string) (minio.ObjectInfo, error) {
	if objectKey == "" {
		return minio.ObjectInfo{}, fmt.Errorf("object_key cannot be empty")
	}

	object, err := c.internalMinio.StatObject(ctx, c.bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		minioErr := minio.ToErrorResponse(err)
		if minioErr.Code == minio.NoSuchKey {
			return minio.ObjectInfo{}, ErrObjectNotFound
		}
		return minio.ObjectInfo{}, err
	}

	return object, nil
}
