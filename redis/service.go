package redis

import (
	"crypto/tls"
	"github.com/deng00/go-base/cache/redis"
	redisV8 "github.com/go-redis/redis/v8"
)

func NewRedis(redisConfig *redis.Config) *redisV8.Client {
	//redis init
	config := &redisV8.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Pass, // no password set
		DB:       redisConfig.DB,   // use default DB
		PoolSize: redisConfig.PoolSize,
	}
	if redisConfig.TlsSkipVerify == true {
		config.TLSConfig = &tls.Config{InsecureSkipVerify: redisConfig.TlsSkipVerify}
	}
	_redisClient := redisV8.NewClient(config)

	return _redisClient
}
