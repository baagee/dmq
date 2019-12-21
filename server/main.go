package main

import (
	"github.com/baagee/dmq/common"
	"log"
)

func main() {
	//bb, _ := json.Marshal(common.Config)
	//log.Println(string(bb))

	log.SetFlags(log.LstdFlags | log.Llongfile)
	app := App{
		msgDetailChan: make(chan common.Message, common.Config.MsgDetailChanLen),
		msgPointChan:  make(chan string, common.Config.MsgPointChanLen),
		msgBucketChan: make(chan string, common.Config.MsgBucketChanLen),
	}
	app.GetPointFromRedis()
}
