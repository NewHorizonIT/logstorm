package redis

import (
	"context"
	"encoding/json"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *goredis.Client
}

func NewRedisCache(client *goredis.Client) Cache {
	return &RedisCache{
		client: client,
	}
}

func (r *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)

	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string, dest any) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {

	return r.client.Del(ctx, key).Err()
}
