package main

import (
	"dmq/client/handle"
	"dmq/redis"
	"dmq/util"
	"fmt"
	"net/http"
)

func main() {
	err, args := getArgs()
	if err != nil {
		fmt.Println("Error: " + err.Error())
	} else {
		redisConfig, err := util.GetRedisConfig(args["config"].(string))
		if err != nil {
			fmt.Println("Error: " + err.Error())
		} else {
			redis.InitPool(redisConfig)
			http.HandleFunc("/send/single", handle.Single)
			err = http.ListenAndServe(fmt.Sprintf(":%d", args["port"].(int)), nil)
			if err != nil {
				fmt.Println("Error: " + err.Error())
			}
		}
		//http.HandleFunc("/send/multiple", send)
	}
}

func getArgs() (error, map[string]interface{}) {
	args, err := util.ParseArgs()
	return err, args
}
