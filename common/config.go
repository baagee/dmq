package common

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	//全局变量 Config配置信息
	Config baseConfig
)

//基本配置
type baseConfig struct {
	HttpPort             uint            `yaml:"http_port"`
	ProductList          []ProductConfig `yaml:"product_list"`
	MsgSaveDays          uint            `yaml:"msg_save_days"`
	GetPointSleep        uint            `yaml:"get_point_sleep"`
	MsgPointChanLen      uint            `yaml:"msg_point_chan_len"`
	MsgBucketChanLen     uint            `yaml:"msg_bucket_chan_len"`
	MsgDetailChanLen     uint            `yaml:"msg_detail_chan_len"`
	BucketCoroutineCount uint            `yaml:"bucket_coroutine_count"`
	MsgCoroutineCount    uint            `yaml:"msg_coroutine_count"`
	WorkPoolSize         uint            `yaml:"work_pool_size"`
	DisableSuccessLog    uint            `yaml:"disable_success_log"`
	RequestTraceIdKey    string          `yaml:"request_trace_id_key"`
	Redis                redisConfig     `yaml:"redis"`
	CommandMap           map[string]commandConfig
	CommandFileList      []string `yaml:"command_list"`
}

//redis配置信息
type redisConfig struct {
	Host            string `yaml:"host"`
	Port            uint   `yaml:"port"`
	Password        string `yaml:"password"`
	Db              uint   `yaml:"db"`
	PoolSize        uint   `yaml:"pool_size"`
	MaxRetries      uint   `yaml:"max_retries"`
	MinRetryBackoff uint   `yaml:"min_retry_backoff"`
	MaxRetryBackoff uint   `yaml:"max_retry_backoff"`
	ReadTimeout     uint   `yaml:"read_timeout"`
	WriteTimeout    uint   `yaml:"write_timeout"`
	MinIdleConns    uint   `yaml:"min_idle_conns"`
	PoolTimeout     uint   `yaml:"pool_timeout"`
}

// command配置
type commandConfig struct {
	Project      string           `yaml:"project"`
	Command      string           `yaml:"command"`
	ConsumerList []ConsumerConfig `yaml:"consumer_list"`
}

// 生产者
type ProductConfig struct {
	Project string   `yaml:"project"`
	AllowIp []string `yaml:"allow_ip"`
}

// 消费者配置
type ConsumerConfig struct {
	Name       string `yaml:"name"`
	Host       string `yaml:"host"`
	Path       string `yaml:"path"`
	Timeout    uint   `yaml:"timeout"`
	RetryTimes uint   `yaml:"retry_times"`
	Interval   uint   `yaml:"interval"`
}

// 解析配置文件
func init() {
	if len(os.Args) < 2 {
		//传入配置文件路径
		ExitWithNotice(ThrowNotice(ErrorCodeDefault, errors.New("请输入配置文件路径")))
	}
	configPath, err := filepath.Abs(os.Args[1])
	if err != nil {
		ExitWithNotice(ThrowNotice(ErrorCodeDefault, err))
	}
	baseConfigFile, err := filepath.Abs(configPath + "/" + "config.yaml")
	if err != nil {
		ExitWithNotice(ThrowNotice(ErrorCodeDefault, err))
	}
	if FileExists(baseConfigFile) == false {
		ExitWithNotice(ThrowNotice(ErrorCodeDefault, errors.New("config.yaml配置文件不存在")))
	}

	err = GetYamlConfig(baseConfigFile, &Config)
	if err != nil {
		ExitWithNotice(ThrowNotice(ErrorCodeDefault, err))
	}
	var (
		cmdConfigFile string
		commandConf   commandConfig
	)
	Config.CommandMap = make(map[string]commandConfig, len(Config.CommandFileList))
	for _, cFile := range Config.CommandFileList {
		cmdConfigFile = configPath + "/" + cFile
		if FileExists(cmdConfigFile) == false {
			ExitWithNotice(ThrowNotice(ErrorCodeDefault, errors.New(fmt.Sprintf("%s 文件不存在", cFile))))
		}
		err := GetYamlConfig(cmdConfigFile, &commandConf)
		if err != nil {
			ExitWithNotice(ThrowNotice(ErrorCodeDefault, err))
		}
		Config.CommandMap[GetConfigCmdKey(commandConf.Command)] = commandConf
	}
	log.Println("config init success")
}
