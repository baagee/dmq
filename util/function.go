package util

import (
	"errors"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

func GetFullBucket(point string, bucket string) string {
	key := fmt.Sprintf("%s:%s", point, bucket)
	return key
}

func GetPointKey(project string) string {
	mainKey := fmt.Sprintf("mmq:%s:points", project)
	return mainKey
}

func GetNumberId() uint64 {
	nano := time.Now().UnixNano()
	return uint64(nano + rand.Int63n(8999) + 1000)
}

//解析命令行参数
func ParseArgs() (map[string]interface{}, error) {
	//默认9090端口
	port := 9090
	// 默认当前目录的配置文件
	config := "redis.yaml"
	flag.IntVar(&port, "p", port, "-p port")
	flag.StringVar(&config, "c", config, "-c redis.yaml")
	flag.Parse()
	idx := strings.Index(config, "/")
	file := ""
	if idx > 0 {
		// 相对路径
		pwdDir, _ := os.Getwd()
		file = pwdDir + "/" + config
	} else {
		// 全路径
		file = config
	}
	ret := make(map[string]interface{})
	if !FileExists(file) {
		return ret, errors.New(fmt.Sprintf("%s 文件不存在\n", file))
	} else {
		ret["port"] = port
		ret["config"] = file
		return ret, nil
	}
}

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

//解析配置文件信息
func GetYamlConfig(configFile string) (map[string]string, error) {
	bytes, err := ioutil.ReadFile(configFile)
	mc := make(map[string]string)
	if err != nil {
		return mc, err
	} else {
		err := yaml.Unmarshal(bytes, &mc)
		if err != nil {
			return mc, err
		} else {
			return mc, nil
		}
	}
}

func GetRedisConfig(file string) (map[string]string, error) {
	config := make(map[string]string)
	config, err := GetYamlConfig(file)
	if err != nil {
		return map[string]string{}, err
	}
	return config, nil
}
