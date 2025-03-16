package event_fsm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type rClient struct {
	rdb redis.UniversalClient
}

func newRClient(rdb redis.UniversalClient) *rClient {
	return &rClient{
		rdb: rdb,
	}
}

func (c *rClient) Get(ctx context.Context, key string, value any) error {
	v, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return fmt.Errorf("rClient.Get: %w", err)
	}

	if err := json.Unmarshal(v, value); err != nil {
		return fmt.Errorf("rClient.Get: %w", err)
	}

	return nil
}

func (c *rClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	v, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("rClient.Set: %w", err)
	}

	if err := c.rdb.Set(ctx, key, v, ttl).Err(); err != nil {
		return fmt.Errorf("rClient.Set: %w", err)
	}

	return nil
}
