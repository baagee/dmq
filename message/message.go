package message

import (
	"dmq/redis"
	"log"
)

type Message struct {
	Id        uint64
	Project   string
	Cmd       string
	Timestamp uint64
	Params    string
	List      string
}

func (this *Message) Save() {
	// zset value 储存时刻任务列表名字eg:23456754:point score 执行时间戳
	point, err := redis.SavePoint(this.Timestamp, this.Project)
	log.Println("point=" + point)
	if err != nil {
		log.Println(err.Error())
	} else {
		key, err := redis.SaveList(point, this.List)
		if err != nil {
			log.Print(err.Error())
		} else {
			log.Println("list key=" + key)
			// hash 结构储存每个时间点下面每个项目的任务列表详情
			err := redis.SaveDetail(key, this.Cmd, this.Timestamp, this.Params, this.Project, this.Id)
			if err != nil {
				log.Println(err.Error())
			}
		}
	}
}
