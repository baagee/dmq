package message

import (
	"dmq/redis"
	"encoding/json"
)

type Message struct {
	Id        uint64
	Project   string
	Cmd       string
	Timestamp uint64
	Params    string
	Bucket    string
}

func (this *Message) Save() error {
	//保存项目时间点  timestamp:point
	point, err := redis.SavePoint(this.Timestamp, this.Project)
	if err != nil {
		return err
	} else {
		//保存此时间点下的bucket point:bucket
		pointBucket, err := redis.SaveBucket(point, this.Bucket)
		if err != nil {
			return err
		} else {
			jsonBytes, err := json.Marshal(*this)
			if err != nil {
				return err
			} else {
				msgJson := string(jsonBytes)
				// 向这个point:bucket lpush 放详细任务信息
				err := redis.SaveDetail(pointBucket, msgJson)
				if err != nil {
					return err
				} else {
					return nil
				}
			}
		}
	}
}
