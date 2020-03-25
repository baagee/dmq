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
	MessageStatusWaiting = 1 //默认状态
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
	RequestId  string `json:"request_id"`  // 请求ID
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
		msgStr := string(bytes)
		// 每个消息针对每个消费者的状态
		messageStatusMap := make(map[string]interface{}, len(cm.ConsumerList))
		for _, consumer := range cm.ConsumerList {
			// ID=>status 消息状态 hash=msgId:status field=consumer value=status
			messageStatusMap[GetMessageStatusHashField(consumer.Host, consumer.Path)] = MessageStatusWaiting
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

		// 3 将任务hash放入当前时刻当前bucket的任务列表
		messageListName := GetMessageHashListName(bucketName)
		redisTx.LPush(messageListName, m.hash)

		//4 将任务状态保存
		messageStatusHashKey := GetMessageStatusHashName(m.Id)
		redisTx.HMSet(messageStatusHashKey, messageStatusMap)
		// 设置过期时间 2*expireTime
		redisTx.Expire(messageStatusHashKey, 2*expireTime)

		// 5 增加 project Y-m-d-H 按天纬度记录ID和hash 过期时间 expireTime
		//			id=>hash的关系 后面可以根据hash获取消息详情
		id2hashDayKey := GetProjectDayHashKey(m.Project, m.Timestamp)
		id2hashIdFieldKey := GetMessageId2HashFieldKey(m.Id)
		redisTx.HSet(id2hashDayKey, id2hashIdFieldKey, m.hash)
		redisTx.Expire(id2hashDayKey, expireTime) //和下面的详情保持一样的过期时间

		// 6 增加消息标记 便于去重 通过hash可以查看m详情 hash=>m
		// 保存消息全部信息 key=hash value=json_encode(m)
		redisTx.Set(GetMessageDetailKey(m.hash), msgStr, expireTime) //expireTime

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

//获取消息消费状态
func (m *Message) Status() (map[string]string, error) {
	consumerStatus, err := RedisCli.HGetAll(GetMessageStatusHashName(m.Id)).Result()
	if err != nil {
		return map[string]string{}, err
	}
	for consumer, status := range consumerStatus {
		s, err := strconv.Atoi(status)
		if err != nil {
			return map[string]string{}, err
		}
		consumerStatus[consumer] = switchStatus(s)
	}
	return consumerStatus, nil
}

//消息数字状态转化为字符串描述
func switchStatus(status int) string {
	switch status {
	case MessageStatusWaiting:
		return "waiting"
	case MessageStatusDoing:
		return "doing"
	case MessageStatusDone:
		return "done"
	case MessageStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

//检查消息是否存在 存在就返回已存在的消息ID
func (m *Message) CheckExists() uint64 {
	m.hash = StringHash(fmt.Sprintf("%s:%d:%s:%s:%s", m.Cmd, m.Timestamp, m.Params, m.Project, m.Bucket))
	ret, err := RedisCli.Get(GetMessageDetailKey(m.hash)).Result()
	if err != nil {
		if err != redis.Nil {
			RecordError(err)
		}
		// 不存在
		return 0
	}

	err = json.Unmarshal([]byte(ret), m)
	if err != nil {
		RecordError(err)
		return 0
	}
	return m.Id
}

////获取最近的时间点
//func (m *Message) GetTimePoint() (redis.Z, error) {
//	zRes, err := RedisCli.ZRangeWithScores(GetPointGroup(m.Project), 0, 0).Result()
//	if err != nil {
//		return redis.Z{}, ThrowNotice(ErrorCodeFoundPointFailed, err)
//	}
//	if len(zRes) == 1 {
//		return zRes[0], nil
//	}
//	return redis.Z{}, nil
//}

// 获取最近的时间点并删除 lua script 保证原子性
func (m *Message) GetTimePoint() (string, error) {
	var luaScript = `
local key = KEYS[1]
-- max score
local _end= ARGV[1]

local res=redis.call('ZRANGEBYSCORE', key, 0, _end, 'WITHSCORES', 'LIMIT', 0, 1)
if (#res == 0) then
-- empty return 0
    return 0
else
    if (redis.call('ZREM', key, res[1]) == 1) then
    -- rem success return res
        return res
    else
    -- rem failed return false
        return false
    end
end
`
	cmdSha, err := RedisCli.ScriptLoad(luaScript).Result()
	if err != nil {
		return "", ThrowNotice(ErrorCodeFoundPointFailed, err)
	}
	zRes, err := RedisCli.EvalSha(cmdSha, []string{GetPointGroup(m.Project)}, time.Now().Unix()).Result()
	if err != nil {
		return "", ThrowNotice(ErrorCodeFoundPointFailed, err)
	}
	if zRes == false {
		return "", ThrowNotice(ErrorCodeFoundPointFailed, errors.New("lua: point delete failed"))
	}
	if zRes == int64(0) {
		return "", nil
	}
	//转化为string
	point := zRes.([]interface{})[0].(string)
	return point, nil
}

////删除时间点
//func (m *Message) RemoveTimePoint(point string) (bool, error) {
//	remCount, err := RedisCli.ZRem(GetPointGroup(m.Project), point).Result()
//	if err != nil {
//		return false, ThrowNotice(ErrorCodeRemovePointFailed, err)
//	}
//	if remCount > 0 {
//		// 如果真删除了返回true
//		return true, nil
//	}
//	// 没有删除 可能已经被删除了
//	return false, ThrowNotice(ErrorCodeRemovePointFailed, errors.New(fmt.Sprintf("时间点[%s]已被清除", point)))
//}

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
func (m *Message) GetBucketMessages(bucket string) []Message {
	var detailList []Message
	// 获取bucket list的所有hash
	messageHashListName := GetMessageHashListName(bucket)
	hashList, err := RedisCli.LRange(messageHashListName, 0, -1).Result()
	if err != nil {
		RecordError(err)
		return detailList
	}
	// 批量通过hash获取消息
	var hashKeyList []string
	for _, h := range hashList {
		hashKeyList = append(hashKeyList, GetMessageDetailKey(h))
	}
	msgStrList, err2 := RedisCli.MGet(hashKeyList...).Result()
	if err2 != nil {
		RecordError(err)
		return detailList
	}
	for _, msgStr := range msgStrList {
		var newMsg Message
		msgStr2 := msgStr.(interface{}).(string)
		err := json.Unmarshal([]byte(msgStr2), &newMsg)
		if err != nil {
			RecordError(err)
			continue
		}
		detailList = append(detailList, newMsg)
	}
	// 销毁bucket
	RedisCli.Del(messageHashListName).Result()
	return detailList
}

//设置消息消费状态
func (m *Message) SetMessageStatus(host string, path string, status int) {
	field := GetMessageStatusHashField(host, path)
	messageStatusHashKey := GetMessageStatusHashName(m.Id)
	if status == MessageStatusFailed {
		log.Printf("message: %d %s failed", m.Id, field)
	} else if status == MessageStatusDone {
		if Config.DisableSuccessLog == 0 {
			// 不输出消费成功的log
			log.Printf("message: %d %s success", m.Id, field)
		}
	}
	_, err := RedisCli.HSet(messageStatusHashKey, field, status).Result()
	if err != nil {
		RecordError(err)
	}
}
