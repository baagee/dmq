package redis

import (
	"dmq/util"
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

//保存时间点
func SavePoint(timestamp uint64, project string) (string, error) {
	redisCli := pool.Get()
	defer redisCli.Close()
	point := fmt.Sprintf("%d:point", timestamp)
	_, err := redis.Int64(redisCli.Do("zAdd", util.GetPointKey(project), timestamp, point))
	if err != nil {
		return "", err
	}
	return point, nil
}

//相同bucket的消息储存在一个列表
func SaveBucket(point string, bucket string) (string, error) {
	key := util.GetFullBucket(point, bucket)
	redisCli := pool.Get()
	defer redisCli.Close()
	_, err := redis.Int64(redisCli.Do("sAdd", point, bucket))
	if err != nil {
		return "", err
	}
	return key, nil
}

//保存消息详情
func SaveDetail(pointBucket string, msgJson string) error {
	redisCli := pool.Get()
	defer redisCli.Close()

	//向list放任务详情 同一个时间点 同一个list类型的任务详情
	_, err := redis.Int64(redisCli.Do("lPush", pointBucket, msgJson))
	if err != nil {
		return err
	} else {
		//当前队列的长度
		return nil
	}
}

//获取最近的时间点
func GetPoint(project string) (string, error) {
	redisCli := pool.Get()
	defer redisCli.Close()
	mainKey := util.GetPointKey(project)
	res, err := redis.Strings(redisCli.Do("zRange", mainKey, 0, 0, "WITHSCORES"))
	if err != nil {
		return "", err
	} else {
		if len(res) >= 1 {
			timePoint, _ := strconv.ParseInt(res[1], 10, 64)
			curTime := time.Now().Unix()
			if curTime >= timePoint {
				// 删除
				redisCli.Do("zRem", mainKey, res[0])
				return res[0], nil
			} else {
				return "", nil
			}
		} else {
			return "", nil
		}
	}
}

// 获取时间点对应的bucket列表
func GetBucket(point string) ([]string, error) {
	redisCli := pool.Get()
	defer redisCli.Close()
	list, err := redis.Strings(redisCli.Do("sMembers", point))
	if err != nil {
		return []string{}, err
	} else {
		//删除这个point
		redisCli.Do("del", point)
		return list, nil
	}
}

//获取bucket对应的任务详情
func GetBucketDetail(bucket string) ([]string, error) {
	redisCli := pool.Get()
	defer redisCli.Close()
	var detailList []string
	for {
		itemDetail, err := redis.String(redisCli.Do("rPop", bucket))
		if err == redis.ErrNil {
			//log.Println("redis rpop读取完毕")
			break
		} else {
			if itemDetail != "" {
				detailList = append(detailList, itemDetail)
			}
		}
	}
	return detailList, nil
}
