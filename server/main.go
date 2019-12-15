package main

import (
	"dmq/redis"
	"dmq/server/handle"
	"dmq/util"
	"fmt"
	"log"
	"time"
)

func main() {
	args, err := util.ParseArgs()
	if err != nil {
		log.Println("Error: " + err.Error())
	} else {
		redisConfig, err := util.GetRedisConfig(args["config"].(string))
		if err != nil {
			fmt.Println("Error: " + err.Error())
		} else {
			redis.InitPool(redisConfig)
			go handle.GetMsgBucket()
			for i := 1; i <= 300; i++ {
				go handle.GetMsgDetail()
			}
			for j := 1; j <= 900; j++ {
				go handle.DoMsgCmd()
			}
			for {
				projects := []string{"svs"}
				for _, project := range projects {
					handle.GetTimePoint(project)
				}
				time.Sleep(time.Millisecond * 400)
			}
		}
	}
}
