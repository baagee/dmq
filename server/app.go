package main

import (
	"github.com/baagee/dmq/common"
	"github.com/baagee/dmq/server/handler"
	"log"
	"time"
)

type App struct {
	msgPointChan  chan string
	msgBucketChan chan string
	msgDetailChan chan common.Message
}

//获取要执行的时间点
func (app *App) GetPointFromRedis() {
	for {
		productList := common.Config.ProductList
		for _, product := range productList {
			member, err := handler.GetTimePoint(product.Project)
			if err != nil {
				common.RecordError(err)
				continue
			}
			if member.Member == nil {
				continue
			}
			point := member.Member.(string)
			score := int64(member.Score)
			log.Println(point, score)
			if time.Now().Unix() < score {
				// 还没到点
				continue
			}

			//到点了 可以执行了 先删除
			err = handler.RemoveTimePoint(product.Project, point)
			if err != nil {
				common.RecordError(err)
				continue
			}
			// 放入pointChan
			app.msgPointChan <- point
		}
		time.Sleep(time.Millisecond * 1000)
	}
}
