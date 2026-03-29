package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/config"
	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	db                    *goredis.Client
	failedCodeAttemptsTTL time.Duration
	codeCaptchaThreshold  int
}

func New(redisConfig config.RedisConfig, captchaConfig config.CaptchaConfig) (*Client, error) {
	opt, err := goredis.ParseURL(redisConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	rdb := goredis.NewClient(opt)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{
		db:                    rdb,
		failedCodeAttemptsTTL: captchaConfig.FailedCodeAttemptsTTL,
		codeCaptchaThreshold:  captchaConfig.CodeCaptchaThreshold,
	}, nil
}

func (c *Client) Close() error {
	err := c.db.Close()
	return err
}
