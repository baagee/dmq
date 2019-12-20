package common

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

//判断文件/目录是否存在
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	} else {
		if os.IsNotExist(err) {
			return false
		}
	}
	return false
}

// 获取配置信息
func GetYamlConfig(configFile string, mc interface{}) error {
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	} else {
		err := yaml.Unmarshal(bytes, mc)
		if err != nil {
			return err
		} else {
			return nil
		}
	}
}

// 输出错误信息并退出
func ExitWithNotice(notice Notice) {
	fmt.Println("Error: " + notice.Error())
	os.Exit(notice.Code())
}

//获取客户端IP
func GetClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

// 基于redis生成数字ID
func GenerateIds(count int64) []uint64 {
	var ids []uint64
	number, err := RedisCli.IncrBy(fmt.Sprintf("%s:%s", RedisKeyPrefix, IdIncrKey), count).Result()
	if err != nil {
		log.Println("GenerateIds err:" + err.Error())
		//降级 使用时间戳
		number = time.Now().Unix()
	}

	var i, id int64
	for i = 0; i < count; i++ {
		id = time.Now().Unix()*10000 + (number-count+i+1)*10000 + rand.Int63n(10000)
		ids = append(ids, uint64(id))
	}
	return ids
}

//获取时间点分组的key 按照项目分组
func GetPointGroup(project string) string {
	return fmt.Sprintf("%s:%s:points", RedisKeyPrefix, project)
}

// 获取消息触发时间点名字
func GetPointName(project string, timestamp uint64) string {
	return fmt.Sprintf("%s:%s:point:%d", RedisKeyPrefix, project, timestamp)
}

// 获取消息bucket名字
func GetBucketName(bucket string, pointName string) string {
	return fmt.Sprintf("%s:%s", pointName, bucket)
}

// 获取储存消息的list名字
func GetMessageListName(bucketName string) string {
	return fmt.Sprintf("%s:messages", bucketName)
}
