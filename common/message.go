package common

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"
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
}

//保存消息
func (m *Message) Save() error {
	bytes, err := json.Marshal(*m)
	if err != nil {
		return ThrowNotice(ErrorCodeJsonMarshal, err)
	}

	consume, exists := Config.CommandMap[GetConfigCmdKey(m.Cmd)]
	//配置文件储存了这个cmd的配置
	if exists {
		pointGroupName := GetPointGroup(m.Project)
		pointName := GetPointName(m.Project, m.Timestamp)
		bucketName := GetBucketName(m.Bucket, pointName)
		messageListName := GetMessageIdListName(bucketName)
		messageStatusHashKey := GetMessageStatusHashName(m.Id)
		messageDetailKey := GetMessageDetailKey(m.Id)

		pointScore := strconv.FormatUint(m.Timestamp, 10)
		expireTimeD := time.Duration(m.Timestamp+3600*24*uint64(Config.MsgSaveDays)-uint64(time.Now().Unix())) * time.Second
		expireTime := strconv.FormatFloat(expireTimeD.Seconds(), 'f', 0, 64)
		msgStr := string(bytes)

		keys := []string{
			pointGroupName, pointName, bucketName, messageListName, messageStatusHashKey, messageDetailKey,
		}
		args := []string{pointScore, strconv.FormatUint(m.Id, 10), expireTime, msgStr}

		// 每个消息针对每个消费者的状态
		for _, consumer := range consume.ConsumerList {
			// ID=>status 消息状态 hash=msgId:status field=consumer value=status
			args = append(args, GetMessageStatusHashField(consumer.Host, consumer.Path))
			args = append(args, strconv.Itoa(MessageStatusWaiting))
		}

		zRes, err := RedisCli.EvalSha(SaveMessageSha, keys, args).Result()
		result := zRes.(interface{}).(int64)
		if err != nil {
			return ThrowNotice(ErrorCodeRedisSave, err)
		}
		if result != int64(1) {
			return ThrowNotice(ErrorCodeRedisSave, errors.New("lua: message save failed"))
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

// 获取最近的时间点并删除 lua script 保证原子性
func (m *Message) GetTimePoint() (string, error) {
	zRes, err := RedisCli.EvalSha(GetTimePointSha, []string{GetPointGroup(m.Project)}, time.Now().Unix()).Result()
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
	messageIdListName := GetMessageIdListName(bucket)
	msgIdList, err := RedisCli.LRange(messageIdListName, 0, -1).Result()
	if err != nil {
		RecordError(err)
		return detailList
	}
	// 批量通过hash获取消息
	var hashKeyList []string
	for _, msgId := range msgIdList {
		idInt, err2 := strconv.ParseInt(msgId, 10, 64)
		if err2 != nil {
			RecordError(err2)
			continue
		}
		hashKeyList = append(hashKeyList, GetMessageDetailKey(uint64(idInt)))
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
	RedisCli.Del(messageIdListName).Result()
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

// 获取消息详情
func (m *Message) GetMessageDetail() error {
	msgJson, err := RedisCli.Get(GetMessageDetailKey(m.Id)).Result()
	if len(msgJson) == 0 {
		return ThrowNotice(ErrorCodeGetMessageFailed, errors.New("没找到对应的消息 msgId:"+strconv.FormatUint(m.Id, 10)))
	}
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(msgJson), m)
	if err != nil {
		return err
	}
	return nil
}
