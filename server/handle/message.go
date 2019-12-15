package handle

import (
	"dmq/message"
	"dmq/redis"
	"dmq/util"
	"encoding/json"
	"log"
)

var (
	msgPointChan  = make(chan string, 1000)
	msgBucketChan = make(chan string, 3000)
	msgDetailChan = make(chan message.Message, 10000)
)

// 获取时间点消息bucket
func GetMsgBucket() {
	for {
		point := <-msgPointChan
		if point != "" {
			log.Println("从msgPointChan获取到point: " + point)
			bucketList, err := redis.GetBucket(point)
			if err != nil {
				log.Println("Error get msg list: " + err.Error())
			} else {
				for _, bucket := range bucketList {
					bucket = util.GetFullBucket(point, bucket)
					log.Printf("%s 添加到msgBucketChan", bucket)
					msgBucketChan <- bucket
				}
			}
		}
	}
}

//获取时间点放入point管道
func GetTimePoint(project string) {
	point, err := redis.GetPoint(project)
	//log.Println(point)
	if err != nil {
		log.Println("Error: GetPoint " + err.Error())
	} else {
		if point != "" {
			log.Println("point=" + point)
			//放入pointChan
			msgPointChan <- point
		}
	}
}

// 获取bucket对应的任务列表
func GetMsgDetail() {
	for {
		itemBucket := <-msgBucketChan
		log.Println("list=" + itemBucket)
		if itemBucket != "" {
			//获取对应的任务详情列表
			detailList, err := redis.GetBucketDetail(itemBucket)
			if err != nil {
				log.Println("Error GetListDetail: " + err.Error())
			} else {
				log.Println("读取" + itemBucket + " 将任务详情放入msgDetailChan")
				for _, detail := range detailList {
					var msg message.Message
					err := json.Unmarshal([]byte(detail), &msg)
					if err != nil {
						log.Println("Error json decode: " + err.Error())
					} else {
						msgDetailChan <- msg
					}
				}
			}
		}
	}
}

//执行任务
func DoMsgCmd() {
	for {
		msg := <-msgDetailChan
		log.Println("执行任务：", msg)
	}
}
