package common

import (
	"crypto/sha1"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
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
	byt, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	} else {
		err := yaml.Unmarshal(byt, mc)
		if err != nil {
			return err
		} else {
			return nil
		}
	}
}

// 输出错误信息并退出
func ExitWithNotice(notice Notice) {
	log.Println("Error: " + notice.Error())
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

//获取生成ID的基数
func GetIdBaseNumber(count int64) int64 {
	number, err := RedisCli.IncrBy(fmt.Sprintf("%s:%s", RedisKeyPrefix, IdIncrKey), count).Result()
	if err != nil {
		RecordError(err)
		//降级 使用时间戳
		number = time.Now().Unix()
	}
	return number
}

// 生成ID
func GenerateId(i int64, number int64, count int64) uint64 {
	id := time.Now().Unix()*10000 + (number-count+i+1)*10000 + rand.Int63n(10000)
	return uint64(id)
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
func GetMessageIdListName(bucketName string) string {
	return fmt.Sprintf("%s:message:id:list", bucketName)
}

//	获取消息状态的hash key 小时区分
func GetMessageStatusHashName(id uint64) string {
	return fmt.Sprintf("%s:message:status:%d", RedisKeyPrefix, id)
}

//	获取消息状态的hash field key
func GetMessageStatusHashField(consumerName string) string {
	return fmt.Sprintf("%s:consumer:%s", RedisKeyPrefix, consumerName)
}

//获取配置信息的cmd map key
func GetConfigCmdKey(cmd string) string {
	return fmt.Sprintf("cmd-%s", strings.Replace(cmd, ":", "-", -1))
}

// 消息是否存在的hash key
func GetMessageDetailKey(msgId uint64) string {
	return fmt.Sprintf("%s:message:detail:%d", RedisKeyPrefix, msgId)
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

//获取字符串hash
func StringHash(data string) string {
	t := sha1.New()
	t.Write([]byte(data))
	return fmt.Sprintf("%x", t.Sum(nil))
}

// 拷贝文件
func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err1 := os.Stat(src)
	if err1 != nil {
		return 0, err1
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err2 := os.Open(src)
	if err2 != nil {
		return 0, err2
	}
	defer source.Close()

	destination, err3 := os.Create(dst)
	if err3 != nil {
		return 0, err3
	}
	defer destination.Close()
	nBytes, err4 := io.Copy(destination, source)
	return nBytes, err4
}

//自动切割Log
func AutoSplitLog(logType string) {
	//切割日志
	ticker := time.NewTicker(time.Second * 3600)
	logFile := fmt.Sprintf("./log/%s.log", logType)
	go func() {
		for {
			<-ticker.C
			if FileExists(logFile) {
				CopyFile(logFile, logFile+time.Now().Format("2006010215")+".log")
				os.Truncate(logFile, 0)
			}
		}
	}()
}
