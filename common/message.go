package common

import (
	"encoding/json"
	"errors"
	"github.com/go-redis/redis"
	"log"
)

const (
	MessageStatusDefault = 1 //默认状态
	MessageStatusDoing   = 2 //正在消费
	MessageStatusDone    = 3 //消费完
	MessageStatusFailed  = 4 //消费失败
)

type Message struct {
	Id         uint64 `json:"id"`          // 消息ID
	Cmd        string `json:"cmd"`         // command
	Timestamp  uint64 `json:"timestamp"`   // 执行时间
	Params     string `json:"params"`      // 命令参数
	Project    string `json:"project"`     // 项目
	Bucket     string `json:"bucket"`      // 消息桶
	CreateTime uint64 `json:"create_time"` // 创建时间
}

//保存消息
func (m *Message) Save() error {
	// TODO 消息去重
	log.Printf("save: %+v\n", *m)
	bytes, err := json.Marshal(*m)
	if err != nil {
		return ThrowNotice(ErrorCodeJsonMarshal, err)
	}

	cm, exists := Config.CommandMap[GetConfigCmdKey(m.Cmd)]
	//配置文件储存了这个cmd的配置
	if exists {
		// 每个消息针对每个消费者的状态
		messageStatusMap := make(map[string]interface{}, len(cm.ConsumerList))
		for _, consumer := range cm.ConsumerList {
			messageStatusMap[GetMessageStatusHashField(m.Id, consumer.Host, consumer.Path)] = MessageStatusDefault
		}

		// 开始redis事务
		redisTx := RedisCli.TxPipeline()
		// 1 保存point
		pointName := GetPointName(m.Project, m.Timestamp)
		pointGroupName := GetPointGroup(m.Project)
		redisTx.ZAdd(pointGroupName, redis.Z{
			Score:  float64(m.Timestamp),
			Member: pointName,
		})
		// 2 增加时间点的bucket
		bucketName := GetBucketName(m.Bucket, pointName)
		redisTx.SAdd(pointName, bucketName)
		// 3 将任务放入当前时刻当前bucket的任务列表
		messageListName := GetMessageListName(bucketName)
		redisTx.LPush(messageListName, string(bytes))
		//4 将任务状态保存
		messageStatusHashKey := GetMessageStatusHashName(m.Timestamp, m.Project)
		redisTx.HMSet(messageStatusHashKey, messageStatusMap)
		//提交
		_, err = redisTx.Exec()
		if err != nil {
			return ThrowNotice(ErrorCodeRedisSave, err)
		}
	} else {
		return ThrowNotice(ErrorCodeRedisSave, errors.New("不合法的cmd"))
	}
	return nil
}
