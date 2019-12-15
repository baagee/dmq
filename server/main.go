package main

import (
	"dmq/redis"
	"log"
	"time"
)

var (
	msgPointChan  = make(chan string, 1000)
	msgListChan   = make(chan string, 3000)
	msgDetailChan = make(chan string, 10000)
)

func main() {
	go getMsgList()
	for i := 1; i <= 30; i++ {
		go getMsgDetail()
	}
	for j := 1; j <= 90; j++ {
		go doMsgCmd()
	}
	for {
		getTimePoint()
		time.Sleep(time.Millisecond * 900)
	}
}

func getTimePoint() {
	point, err := redis.GetPoint("svs")
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

func getMsgList() {
	for {
		point := <-msgPointChan
		//log.Println("point=" + point)
		if point != "" {
			log.Println("从msgPointChan获取到point: " + point)
			list, err := redis.GetList(point)
			if err != nil {
				log.Println("Error get msg list: " + err.Error())
			} else {
				for _, item := range list {
					log.Printf("%s 添加到msgListChan", item)
					msgListChan <- item
				}
			}
		}
	}
}

func getMsgDetail() {
	for {
		itemList := <-msgListChan
		log.Println("list=" + itemList)
		if itemList != "" {
			//获取对应的任务详情列表
			detailList, err := redis.GetListDetail(itemList)
			if err != nil {
				log.Println("Error GetListDetail: " + err.Error())
			} else {
				//log.Println(detailList)
				for _, detail := range detailList {
					log.Println("读取" + itemList + " 将任务详情放入msgDetailChan")
					msgDetailChan <- detail
				}
			}
		}
	}
}

func doMsgCmd() {
	for {
		detail := <-msgDetailChan
		log.Println(detail)
	}
}
