package redis

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
	"strconv"
	"time"
)

var pool *redis.Pool

func init() {
	//初始化redis连接池
	pool = &redis.Pool{
		MaxIdle:     160, //池中最大空闲连接数
		MaxActive:   920, //在给定时间池分配的最大连接数
		IdleTimeout: 120, //在此时间保持空闲后关闭连接
		Dial: func() (redis.Conn, error) {
			//连接redis
			return redis.Dial("tcp", "127.0.0.1:6379")
		},
	}
	log.Println("redis pool init")
}

func SavePoint(timestamp uint64, project string) (string, error) {
	redisCli := pool.Get()
	defer redisCli.Close()
	point := fmt.Sprintf("%s:%d:point", project, timestamp)
	res, err := redis.Int64(redisCli.Do("zAdd", getPointKey(project), timestamp, point))
	if err != nil {
		return "", err
	} else {
		log.Println("savePOint:", res)
	}
	return point, nil
}

//相同list的消息储存在一个列表
func SaveList(point string, list string) (string, error) {
	key := GetFullList(point, list)
	redisCli := pool.Get()
	defer redisCli.Close()
	res, err := redis.Int64(redisCli.Do("sAdd", point, list))
	if err != nil {
		return "", err
	} else {
		log.Println(key)
		log.Println("save list :", res)
	}
	return key, nil
}

func GetFullList(point string, list string) string {
	key := fmt.Sprintf("%s:%s", point, list)
	return key
}

func SaveDetail(key string, cmd string, timestamp uint64, params string, project string, id uint64) error {
	//key := fmt.Sprintf("%s:%s", point, list)
	redisCli := pool.Get()
	defer redisCli.Close()
	msg := make(map[string]string)
	msg = map[string]string{
		"project":   project,
		"cmd":       cmd,
		"timestamp": strconv.FormatUint(timestamp, 10),
		"params":    params,
		"id":        strconv.FormatUint(id, 10),
	}
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		log.Println(err.Error())
		return err
	} else {
		log.Println(string(jsonBytes))
		//向list放任务详情 同一个时间点 同一个list类型的任务详情
		res, err := redis.Int64(redisCli.Do("lPush", key, string(jsonBytes)))
		if err != nil {
			//log.Println(err.Error())
			return err
		} else {
			//当前队列的长度
			log.Println("SaveDetail: ", res)
			return nil
		}
	}
}

func GetPoint(project string) (string, error) {
	redisCli := pool.Get()
	defer redisCli.Close()
	mainKey := getPointKey(project)
	res, err := redis.Strings(redisCli.Do("zRange", mainKey, 0, 0, "WITHSCORES"))
	if err != nil {
		return "", err
	} else {
		//log.Println(res)
		if len(res) >= 1 {
			timePoint, _ := strconv.ParseInt(res[1], 10, 64)
			curTime := time.Now().Unix()
			//log.Println(timePoint, curTime)
			if curTime >= timePoint {
				log.Println("到点可以执行了")
				// 删除
				ef, _ := redis.Int(redisCli.Do("zRem", mainKey, res[0]))
				log.Println("删除point结果：", ef)
				return res[0], nil
			} else {
				log.Println("还未到时间：", res[1])
				return "", nil
			}
		} else {
			return "", nil
		}
	}
}

func GetList(point string) ([]string, error) {
	redisCli := pool.Get()
	defer redisCli.Close()
	list, err := redis.Strings(redisCli.Do("sMembers", point))
	if err != nil {
		return []string{}, err
	} else {
		//log.Println(list)
		//删除这个point
		ef, _ := redis.Int(redisCli.Do("del", point))
		log.Println("删除pointList结果：", ef)
		for i, v := range list {
			list[i] = GetFullList(point, v)
		}
		return list, nil
	}
}

func getPointKey(project string) string {
	mainKey := fmt.Sprintf("mmq:%s:points", project)
	return mainKey
}

func GetListDetail(list string) ([]string, error) {
	//log.Println(list)
	redisCli := pool.Get()
	defer redisCli.Close()
	var detailList []string
	for {
		itemDetail, err := redis.String(redisCli.Do("rPop", list))
		if err == redis.ErrNil {
			log.Println("redis rpop读取完毕")
			break
		} else {
			if itemDetail != "" {
				detailList = append(detailList, itemDetail)
			}
		}
	}
	return detailList, nil
}
