package redis

import (
	"github.com/NewHorizonIT/logstorm/internal/config"
	goredis "github.com/redis/go-redis/v9"
)

func NewClient(conf config.RedisConfig) *goredis.Client {
	client := goredis.NewClient(&goredis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
	})

	return client
}
