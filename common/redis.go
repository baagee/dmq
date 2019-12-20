package common

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
)

var (
	RedisCli *redis.Client
)

func init() {
	//连接redis
	RedisCli = redis.NewClient(&redis.Options{
		Addr:       fmt.Sprintf("%s:%d", Config.Redis.Host, Config.Redis.Port),
		Password:   Config.Redis.Password,
		DB:         int(Config.Redis.Db),
		PoolSize:   int(Config.Redis.PoolSize),
		MaxRetries: int(Config.Redis.MaxRetries),
	})
	if _, err := RedisCli.Ping().Result(); err != nil {
		panic("connect redis error:" + err.Error())
	}
	log.Println("connect redis success")
}
