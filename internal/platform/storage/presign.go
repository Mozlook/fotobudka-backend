package storage

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

func (c *Client) PresignedPutObject(ctx context.Context, objectKey string, expires time.Duration) (*url.URL, error) {
	if objectKey == "" {
		return nil, fmt.Errorf("object_key cannot be empty")
	}

	if expires < time.Second {
		return nil, fmt.Errorf("duration must be longer than 1 second")
	}

	putURL, err := c.publicMinio.PresignedPutObject(ctx, c.bucketName, objectKey, expires)
	if err != nil {
		return nil, fmt.Errorf("presign put object: %w", err)
	}

	return putURL, nil
}
