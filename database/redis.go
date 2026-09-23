package database

import (
	"os"
	"sonla/transcode/util"

	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

func ConnectRedis() (*redis.Client, error) {
	var rd = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	if pong, err := rd.Ping().Result(); err != nil {
		util.Logger.Fatal("Redis: connect failed", zap.Error(err))
		return nil, err
	} else {
		util.Logger.Info("Redis: connect success", zap.String("pong", pong))
		return rd, nil
	}
}
