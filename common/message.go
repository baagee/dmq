package common

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"log"
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
	log.Printf("save: %+v\n", *m)
	bytes, err := json.Marshal(*m)
	if err != nil {
		return ThrowNotice(ErrorCodeJsonMarshal, err)
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
	//提交
	_, err = redisTx.Exec()
	if err != nil {
		return ThrowNotice(ErrorCodeRedisSave, err)
	}
	return nil
}
