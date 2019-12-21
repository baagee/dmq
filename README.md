# dmq
golang 基于redis实现的消息队列，支持延时消息，使用了协程提高并发能力

## 分为client和server端

### client端负责接收消息，提供http接口
`/api/message/single`接口为提交一个消息
```json
{
    "project":"sds", // 所属项目
    "cmd":"sds:create:order",// 消息执行的命令
    "timestamp":0,// 执行的时间 为0立即执行
    "params":"{\"order_id\":34256}",// 消息参数
    "bucket":"order:342543"// 命令bucket
}
```

`/api/message/batch`接口为批量提交消息
```json
[
	{
		"project":"sds", // 所属项目
		"cmd":"sds:create:order",// 消息执行的命令
		"timestamp":0,// 执行的时间 为0立即执行
		"params":"{\"order_id\":34256}",// 消息参数
		"bucket":"order:342543"// 命令bucket
	},
	{
		"project":"svs",
		"cmd":"svs:notice:user",
		"params":"{\"user_id\":34256}",
		"bucket":"notice:1234"
	}
]
```
### server端负责处理消息，到时间点时让消费者处理消息

## 使用
config目录是配置目录，具体看配置文件注释

假设已经编译好执行文件server和client

启动server `path/to/server ../config(相对路径，client的配置文件目录)`

启动client `path/to/client ../config(相对路径，client的配置文件目录)`