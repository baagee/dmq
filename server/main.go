package main

import (
	"dmq/server/handle"
	"time"
)

func main() {
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
