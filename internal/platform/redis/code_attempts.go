package redis

import (
	"context"
	"fmt"
	"strconv"

	goredis "github.com/redis/go-redis/v9"
)

func (c *Client) RegisterFailedCodeAttempt(ctx context.Context, ip string) (int64, error) {
	key := fmt.Sprintf("code_attempts_%s", ip)

	result, err := c.db.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if result == 1 {
		_, err = c.db.Expire(ctx, key, c.failedCodeAttemptsTTL).Result()
		if err != nil {
			return 0, err
		}
	}

	return result, nil
}

func (c *Client) RequiresCodeCaptcha(ctx context.Context, ip string) (bool, error) {
	key := fmt.Sprintf("code_attempts_%s", ip)

	result, err := c.db.Get(ctx, key).Result()
	if err == goredis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	failedAttempts, err := strconv.Atoi(result)
	if err != nil {
		return false, err
	}

	if failedAttempts >= c.codeCaptchaThreshold {
		return true, nil
	}

	return false, nil
}

func (c *Client) ClearFailedCodeAttempts(ctx context.Context, ip string) error {
	key := fmt.Sprintf("code_attempts_%s", ip)

	_, err := c.db.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}
