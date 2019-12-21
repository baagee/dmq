package common

import (
	"bytes"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
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
		RecordError(err)
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

//	获取消息状态的hash key 小时区分
func GetMessageStatusHashName(timestamp uint64, project string) string {
	hour := time.Unix(int64(timestamp), 0).Format("2006-01-02-15")
	return fmt.Sprintf("%s:%s:%s:message:status", RedisKeyPrefix, project, hour)
}

//	获取消息状态的hash field key
func GetMessageStatusHashField(msgId uint64, host string, path string) string {
	return fmt.Sprintf("%s:%d:%s%s", RedisKeyPrefix, msgId, host, path)
}

//获取配置信息的cmd map key
func GetConfigCmdKey(cmd string) string {
	return fmt.Sprintf("cmd-%s", strings.Replace(cmd, ":", "-", -1))
}

// 记录错误信息
func RecordError(err error) {
	switch n := err.(type) {
	case Notice:
		log.Printf("Notice Error：[%d] %s", n.Code(), n.Error())
	default:
		log.Printf("Error: %s", n.Error())
	}
}

// http post请求
func HttpPost(url string, params string, timeout uint) error {
	var jsonBytes = []byte(params)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return ThrowNotice(ErrorCodePreRequestFailed, err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
	resp, err := client.Do(req)
	if err != nil {
		//if strings.Contains(err.Error(), "Client.Timeout exceeded") {
		//	fmt.Println("HTTP post timeout")
		//}
		return ThrowNotice(ErrorCodeRequestFailed, err)
	}
	defer resp.Body.Close()

	//body, _ := ioutil.ReadAll(resp.Body)
	//log.Println("response Body:", string(body))
	if resp.StatusCode != 200 {
		return ThrowNotice(ErrorCodeResponseCodeNot200, errors.New("response code!=200"))
	}
	return nil
}

// 获取消费者下游接口
func GetConsumerFullUrl(host string, path string, msgId uint64) string {
	urlStr := fmt.Sprintf("%s%s", host, path)
	if strings.Index(urlStr, "http") == -1 {
		//不是http开头的 加上http
		urlStr = fmt.Sprintf("http://%s", urlStr)
	}
	// 判断是否已经有参数了
	var rawQuery bool
	u, err := url.Parse(urlStr)
	// 加上消息ID
	if err != nil {
		RecordError(err)
		rawQuery = strings.Index(urlStr, "?") != -1
	} else {
		rawQuery = u.RawQuery != ""
	}
	if rawQuery {
		// 原来有参数
		urlStr += fmt.Sprintf("&msg_id=%d", msgId)
	} else {
		urlStr += fmt.Sprintf("?msg_id=%d", msgId)
	}
	return urlStr
}
