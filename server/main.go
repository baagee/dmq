package main

import (
	"github.com/baagee/dmq/common"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	app := App{
		msgDetailChan: make(chan common.Message, 1000),
		msgPointChan:  make(chan string, 200),
		msgBucketChan: make(chan string, 500),
	}
	app.GetPointFromRedis()
}
