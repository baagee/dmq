package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"strconv"
	"time"
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
	hash       string // 消息体唯一标示 下游获取不到 因为json encode时就没有了
}

//保存消息
func (m *Message) Save() error {
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
		// 消息标记过期时间 从现在到消息的执行时间后n天 这段时间不允许重复
		expireTime := time.Duration(m.Timestamp+3600*24*uint64(Config.MsgNoRepeatDay)-uint64(time.Now().Unix())) * time.Second
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
		// 5 增加消息标记 便于去重
		redisTx.Set(GetMsgRedisFlagKey(m.hash), m.Id, expireTime)
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

//检查消息是否存在 存在就返回已存在的消息ID
func (m *Message) CheckExists() uint64 {
	m.hash = StringHash(fmt.Sprintf("%s:%d:%s:%s:%s", m.Cmd, m.Timestamp, m.Params, m.Project, m.Bucket))
	ret, err := RedisCli.Get(GetMsgRedisFlagKey(m.hash)).Result()
	if err != nil {
		// 不存在
		return 0
	}
	beforeId, err := strconv.Atoi(ret)
	if err != nil {
		return 0
	}
	return uint64(beforeId)
}

//获取最近的时间点
func (m *Message) GetTimePoint() (redis.Z, error) {
	zRes, err := RedisCli.ZRangeWithScores(GetPointGroup(m.Project), 0, 0).Result()
	if err != nil {
		return redis.Z{}, ThrowNotice(ErrorCodeFoundPointFailed, err)
	}
	if len(zRes) == 1 {
		return zRes[0], nil
	}
	return redis.Z{}, nil
}

//删除时间点
func (m *Message) RemoveTimePoint(point string) error {
	_, err := RedisCli.ZRem(GetPointGroup(m.Project), point).Result()
	if err != nil {
		return ThrowNotice(ErrorCodeRemovePointFailed, err)
	}
	return nil
}

// 获取时间点的buckets
func (m *Message) GetPointBuckets(point string) ([]string, error) {
	buckets, err := RedisCli.SMembers(point).Result()
	if err != nil {
		return []string{}, ThrowNotice(ErrorCodeFoundBucketsFailed, err)
	}
	// 删除它
	_, err = RedisCli.Del(point).Result()
	if err != nil {
		return []string{}, ThrowNotice(ErrorCodeRemoveBucketsFailed, err)
	}
	return buckets, nil
}

// 从redis获取bucket对应的任务
func (m *Message) GetBucketMessages(bucket string) []string {
	var detailList []string
	messageListName := GetMessageListName(bucket)
	for {
		itemDetail, err := RedisCli.RPop(messageListName).Result()
		if err != nil {
			if err == redis.Nil {
				//redis rpop读取完毕
				break
			} else {
				RecordError(err)
				continue
			}
		}
		if itemDetail != "" {
			detailList = append(detailList, itemDetail)
		}
	}
	return detailList
}

func (m *Message) SetMessageStatus(host string, path string, status int) {
	field := GetMessageStatusHashField(m.Id, host, path)
	messageStatusHashKey := GetMessageStatusHashName(m.Timestamp, m.Project)
	if status == MessageStatusFailed {
		log.Println(field + " failed")
	} else if status == MessageStatusDone {
		log.Println(field + " success")
	}
	_, err := RedisCli.HSet(messageStatusHashKey, field, status).Result()
	if err != nil {
		RecordError(err)
	}
}
