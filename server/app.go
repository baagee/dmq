package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/baagee/dmq/common"
	"github.com/baagee/dmq/taskpool"
	"log"
	"net/http"
	"strconv"
	"time"
)

type App struct {
	msgPointChan  chan string
	msgBucketChan chan string
	msgDetailChan chan common.Message
	workerPool    *taskpool.Pool
}

//获取要执行的时间点
func (app *App) GetPointFromRedis() {
	productList := common.Config.ProductList
	for {
		for _, product := range productList {
			go app.getProductPoint(product)
		}
		time.Sleep(time.Millisecond * time.Duration(common.Config.GetPointSleep))
	}
}

// 获取当前生产者的时间点
func (app *App) getProductPoint(product common.ProductConfig) {
	msg := common.Message{
		Project: product.Project,
	}
	point, err := msg.GetTimePoint()
	if err != nil {
		common.RecordError(err)
		return
	}
	if point != "" {
		app.msgPointChan <- point
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

//添加消费任务
func (app *App) addConsumerTask(consumer common.ConsumerConfig, msg *common.Message) {
	//log.Println("put consume task into task pool...")
	app.workerPool.AddTask(taskpool.NewTask(func(workId uint) error {
		//log.Println(fmt.Sprintf("workId#%d start runing", workId))
		app.consume(consumer, msg, workId)
		//log.Println(fmt.Sprintf("workId#%d end run", workId))
		return nil
	}))
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
			// 一个协程处理一个消费者 放到工作池中
			app.addConsumerTask(consumer, &msg)
		}
	}
}

//关闭工作池
func (app *App) CloseWorkPool() {
	app.workerPool.Close()
}

//开启工作池
func (app *App) RunWorkPool() {
	go app.workerPool.Run()
}

// 消费
func (app *App) consume(consumer common.ConsumerConfig, msg *common.Message, workId uint) {
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
		err := app.requestConsumer(msg, curUrl, consumer.Timeout, workId)
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
func (app *App) requestConsumer(msg *common.Message, url string, timeout uint, workId uint) error {
	var jsonBytes = []byte(msg.Params)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return common.ThrowNotice(common.ErrorCodePreRequestFailed, err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "dmq(message queue)")
	if msg.RequestId != "" {
		req.Header.Set("X-"+common.Config.RequestTraceIdKey, msg.RequestId) //设置消息生产者请求ID 连贯生产者和消费者
	}
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
	startTime := time.Now()
	resp, err := client.Do(req)

	if err != nil {
		return common.ThrowNotice(common.ErrorCodeRequestFailed, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return common.ThrowNotice(common.ErrorCodeResponseCodeNot200, errors.New("response code!=200"))
	}
	log.Printf("workId [%d] request [%s] cost time:%dms\n", workId, url, time.Now().Sub(startTime)/time.Millisecond)
	return nil
}
