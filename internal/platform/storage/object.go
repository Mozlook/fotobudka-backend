package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

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

func (c *Client) GetObject(ctx context.Context, objectKey string) (*minio.Object, error) {
	if objectKey == "" {
		return nil, fmt.Errorf("object_key cannot be empty")
	}

	object, err := c.internalMinio.GetObject(ctx, c.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		minioErr := minio.ToErrorResponse(err)
		if minioErr.Code == minio.NoSuchKey {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}

	if _, err := object.Stat(); err != nil {
		_ = object.Close()

		minioErr := minio.ToErrorResponse(err)
		if minioErr.Code == minio.NoSuchKey {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}

	return object, nil
}

func (c *Client) PutObject(ctx context.Context, objectKey, contentType string, reader io.Reader, size int64) error {
	if objectKey == "" {
		return fmt.Errorf("object_key cannot be empty")
	}

	if contentType == "" {
		return fmt.Errorf("content_type cannot be empty")
	}

	if size <= 0 {
		return fmt.Errorf("size must be greater than 0")
	}

	_, err := c.internalMinio.PutObject(
		ctx,
		c.bucketName,
		objectKey,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
