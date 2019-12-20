package common

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"time"
)

var (
	RedisCli *redis.Client
)

func init() {
	//连接redis
	RedisCli = redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", Config.Redis.Host, Config.Redis.Port),
		Password:        Config.Redis.Password,
		DB:              int(Config.Redis.Db),
		PoolSize:        int(Config.Redis.PoolSize),
		MaxRetries:      int(Config.Redis.MaxRetries),
		MinRetryBackoff: time.Duration(Config.Redis.MinRetryBackoff) * time.Millisecond,
		MaxRetryBackoff: time.Duration(Config.Redis.MaxRetryBackoff) * time.Millisecond,
		ReadTimeout:     time.Duration(Config.Redis.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(Config.Redis.WriteTimeout) * time.Second,
		MinIdleConns:    int(Config.Redis.MinIdleConns),
		PoolTimeout:     time.Duration(Config.Redis.PoolTimeout) * time.Second,
	})
	if _, err := RedisCli.Ping().Result(); err != nil {
		panic("connect redis error:" + err.Error())
	}
	log.Println("connect redis success")
}
