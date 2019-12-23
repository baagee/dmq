package main

import (
	"github.com/baagee/dmq/common"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	app := App{
		msgDetailChan: make(chan common.Message, common.Config.MsgDetailChanLen),
		msgPointChan:  make(chan string, common.Config.MsgPointChanLen),
		msgBucketChan: make(chan string, common.Config.MsgBucketChanLen),
	}
	// 一个协程获取point时刻的bucket
	go app.GetPointBuckets()

	// 多协程获取时间点的bucket的消息详情
	var i, j uint
	for i = 0; i < common.Config.BucketCoroutineCount; i++ {
		go app.GetBucketMessages()
	}
	// 多协程执行命令
	for j = 0; j < common.Config.MsgCoroutineCount; j++ {
		go app.DoMessageCmd()
	}
	//处理point
	app.GetPointFromRedis()
}
