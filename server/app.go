package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/baagee/dmq/common"
	"log"
	"strings"
	"time"
)

type App struct {
	msgPointChan  chan string
	msgBucketChan chan string
	msgDetailChan chan common.Message
}

//获取要执行的时间点
func (app *App) GetPointFromRedis() {
	productList := common.Config.ProductList
	for {
		for _, product := range productList {
			msg := common.Message{
				Project: product.Project,
			}
			member, err := msg.GetTimePoint()
			if err != nil {
				common.RecordError(err)
				continue
			}
			if member.Member == nil {
				continue
			}
			point := member.Member.(string)
			score := int64(member.Score)
			//log.Println(point, score)
			if time.Now().Unix() < score {
				// 还没到点
				continue
			}

			//到点了 可以执行了 先删除
			err = msg.RemoveTimePoint(point)
			if err != nil {
				common.RecordError(err)
				continue
			}
			// 放入pointChan
			app.msgPointChan <- point
		}
		time.Sleep(time.Millisecond * time.Duration(common.Config.GetPointSleep))
	}
}

// 获取buckets
func (app *App) GetPointBuckets() {
	for {
		point := <-app.msgPointChan
		if point == "" {
			continue
		}
		var msg common.Message
		//log.Println("从msgPointChan获取到point: " + point)
		bucketList, err := msg.GetPointBuckets(point)
		if err != nil {
			common.RecordError(err)
			continue
		}
		if len(bucketList) == 0 {
			continue
		}
		for _, bucket := range bucketList {
			app.msgBucketChan <- bucket
			//log.Printf("%s 添加到msgBucketChan", bucket)
		}
	}
}

//获取bucket对应的消息
func (app *App) GetBucketMessages() {
	for {
		bucket := <-app.msgBucketChan
		if bucket == "" {
			continue
		}
		var msg common.Message
		msgJsonList := msg.GetBucketMessages(bucket)

		for _, jsonStr := range msgJsonList {
			var m common.Message
			err := json.Unmarshal([]byte(jsonStr), &m)
			if err != nil {
				common.RecordError(err)
				continue
			}
			//log.Println("将消息放入detail chan " + m.Cmd)
			app.msgDetailChan <- m
		}
	}
}

// 执行命令
func (app *App) DoMessageCmd() {
	for {
		msg := <-app.msgDetailChan
		consumerList := common.Config.CommandMap[common.GetConfigCmdKey(msg.Cmd)].ConsumerList
		//log.Println(consumerList)
		for _, consumer := range consumerList {
			// 一个协程处理一个消费者
			//requestConsumer(consumer, &msg)
			go requestConsumer(consumer, &msg)
		}
	}
}

// 请求消费者
func requestConsumer(consumer common.ConsumerConfig, msg *common.Message) {
	// 状态设置组成正在做
	msg.SetMessageStatus(consumer.Host, consumer.Path, common.MessageStatusDoing)
	log.Printf("%s%s %d 消费 %d %s\n", consumer.Host, consumer.Path, time.Now().Unix(), msg.Timestamp, msg.Cmd)
	url := fmt.Sprintf("%s%s", consumer.Host, consumer.Path)
	if strings.Index(url, "http") == -1 {
		//不是http开头的 加上http
		url = fmt.Sprintf("http://%s", url)
	}
	var t uint
	success := false
	// 重试机制
	for t = 0; t <= consumer.RetryTimes; t++ {
		err := common.HttpPost(url, msg.Params, consumer.Timeout)
		if err != nil {
			common.RecordError(errors.New(fmt.Sprintf("retry=%d %s", t, err.Error())))
			//稍微休息一下
			time.Sleep(time.Duration(consumer.Interval) * time.Millisecond)
			continue
		}
		success = true
		break
	}
	if success {
		// 成功后改状态为消费成功
		msg.SetMessageStatus(consumer.Host, consumer.Path, common.MessageStatusDone)
	} else {
		// 失败后改状态为消费失败
		msg.SetMessageStatus(consumer.Host, consumer.Path, common.MessageStatusFailed)
	}
}
