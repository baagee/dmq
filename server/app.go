package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/baagee/dmq/common"
	"log"
	"net/http"
	"strconv"
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
	//通过recover保证一个协程失败 不影响其他协程
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Error: %+v", r)
		}
	}()
	for {
		point := <-app.msgPointChan
		if point == "" {
			continue
		}
		var msg common.Message
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
		}
	}
}

//获取bucket对应的消息
func (app *App) GetBucketMessages() {
	//通过recover保证一个协程失败 不影响其他协程
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Error: %+v", r)
		}
	}()
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
			app.msgDetailChan <- m
		}
	}
}

// 执行命令
func (app *App) DoMessageCmd() {
	//通过recover保证一个协程失败 不影响其他协程
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Error: %+v", r)
		}
	}()
	for {
		msg := <-app.msgDetailChan
		consumerList := common.Config.CommandMap[common.GetConfigCmdKey(msg.Cmd)].ConsumerList
		for _, consumer := range consumerList {
			// 一个协程处理一个消费者
			go app.consume(consumer, &msg)
		}
	}
}

// 消费
func (app *App) consume(consumer common.ConsumerConfig, msg *common.Message) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Error: %+v", r)
			// 消息重新进入通道去消费
			app.msgDetailChan <- *msg
		}
	}()
	// 状态设置组成正在做
	msg.SetMessageStatus(consumer.Host, consumer.Path, common.MessageStatusDoing)
	url := common.GetConsumerFullUrl(consumer.Host, consumer.Path, msg.Id)
	var (
		// 重试次数
		retry uint = 0
		//是否请求成功
		success = false
	)
	// 重试机制
	for ; retry <= consumer.RetryTimes; retry++ {
		// 加上重试次数
		curUrl := url + "&retry=" + strconv.FormatUint(uint64(retry), 10)
		err := app.requestConsumer(msg, curUrl, consumer.Timeout)
		if err != nil {
			field := common.GetMessageStatusHashField(consumer.Host, consumer.Path)
			common.RecordError(errors.New(fmt.Sprintf("%s retry=%d %s", field, retry, err.Error())))
			//稍微休息一下
			if retry < consumer.RetryTimes {
				// 最后一次不需要休息了
				time.Sleep(time.Duration(consumer.Interval*(retry+1)) * time.Millisecond)
			}
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

//真正的消费
func (app *App) requestConsumer(msg *common.Message, url string, timeout uint) error {
	var jsonBytes = []byte(msg.Params)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return common.ThrowNotice(common.ErrorCodePreRequestFailed, err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "dmq(message queue)")
	if msg.RequestId != "" {
		req.Header.Set("X-Request-Id", msg.RequestId) //设置消息生产者请求ID 连贯生产者和消费者
	}
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
	resp, err := client.Do(req)
	if err != nil {
		return common.ThrowNotice(common.ErrorCodeRequestFailed, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return common.ThrowNotice(common.ErrorCodeResponseCodeNot200, errors.New("response code!=200"))
	}
	return nil
}
