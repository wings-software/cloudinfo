package modernredis

import (
	"github.com/go-redis/redis"
)

func NewRedisClient(config Config) *redis.Client {
	if config.SentinelEnabled {
		return redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    config.MasterName,
			Password:      config.MasterPassword,
			DB:            config.Db,
			MaxRetries:    3,
			SentinelAddrs: config.SentinelAddress,
		})
	} else {
		return redis.NewClient(&redis.Options{
			Addr:     config.Address,
			Password: config.Password,
			DB:       config.Db,
		})
	}
}
