package runtime

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	Cli *redis.Client
}

func newRedis(rt *Runtime) (*Redis, error) {
	cli := redis.NewClient(&redis.Options{
		Addr:         rt.Config.Redis.Address,
		Username:     rt.Config.Redis.User,
		Password:     rt.Config.Redis.Password,
		DB:           rt.Config.Redis.DB,
		MaxRetries:   rt.Config.Redis.MaxRetries,
		PoolSize:     rt.Config.Redis.PoolSize,
		MinIdleConns: rt.Config.Redis.MinIdle,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := cli.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &Redis{Cli: cli}, nil
}

func (r *Redis) Close() error {
	return r.Cli.Close()
}
