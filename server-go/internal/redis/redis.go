package redis

import (
	"context"
	"os"

	redisv9 "github.com/redis/go-redis/v9"
)

var (
	Client *redisv9.Client
	Ctx    = context.Background()
)

func Init() error {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = "redis://localhost:6379"
	}

	opt, err := redisv9.ParseURL(url)
	if err != nil {
		return err
	}

	Client = redisv9.NewClient(opt)

	if err := Client.Ping(Ctx).Err(); err != nil {
		return err
	}

	return nil
}
