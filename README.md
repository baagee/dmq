# dmq
golang 基于redis实现的消息队列，支持延时消息，使用了协程提高并发能力

## 分为client和server端

### client端负责接收消息，提供http接口

#### POST `/api/message/single`接口为提交一个消息
request:
```json
{
  "project":"sds", // 所属项目 不为空 字符串
  "cmd":"sds:create:order",// 消息执行的命令 不为空 字符串
  "timestamp":0,// 执行的时间点 时间戳 为0立即执行 可为空 int
  "params":"{\"order_id\":34256}",// 消息参数 可为空 字符串
  "request_id":"235423532",// 请求ID(log_id/trace_id) 会通过header带给消费者 可为空 字符串
  "bucket":"order:342543"// 命令bucket 不为空 字符串
}
```
response:
```json
{
  "code": 0,
  "message": "", //如果不是合法的json结构 code>0 message返回错误信息
  "data": 15771179759758 //消息保存成功返回消息ID 失败返回false 参数不合法返回错误信息
}
```

#### POST `/api/message/batch`接口为批量提交消息
request:
```json
[
  {
    "project":"sds", // 所属项目
    "cmd":"sds:create:order",// 消息执行的命令
    "timestamp":0,// 执行的时间点 时间戳 为0立即执行
    "params":"{\"order_id\":34256}",// 消息参数
    "request_id":"235423532",// 请求ID(log_id/trace_id) 会通过header带给消费者(x-request-id) 字符串
    "bucket":"order:342543"// 命令bucket
  },
  {
    "project":"svs",
    "cmd":"svs:notice:user",
    "params":"{\"user_id\":34256}",
    "bucket":"notice:1234"
  },
  {
    "project":"svss",
    "cmd":"svs:notice:user",
    "params":"{\"user_id\":34256}",
    "bucket":"notice:1234"
  }
]
```
response:
```json
{
  "code": 0,
  "message": "",
  "data": [
    15771179759758,//第一个消息保存成功返回消息ID
    false,//第二个消息保存失败返回false
    "不合法的生产者"//第三个消息参数不合法返回错误信息
  ]
}
```

#### GET `/api/message/status?msg_id=15771179759758&consumer=consumer_name`接口为查看消息消费的状态
response:
```json
{
    "code": 0,
    "message": "",
    "data": "done"//waiting/doing/done/failure/unknown
}
```


#### GET `/api/message/detail?msg_id=15771179759758&consumer=consumer_name`接口为查看消息详细信息
response:
```json
{
    "code": 0,
    "message": "",
    "data": {
        "id": 15971219465821,
        "cmd": "ddd:create:order",
        "timestamp": 1597121929,
        "params": "{\"user_id\":88}",
        "project": "sds",
        "bucket": "sds_bucket_5",
        "create_time": 1597121930,
        "request_id": ""
    }
}
```

#### GET `/api/message/pending?consumer=consumer_name&start=123456&end=987654`接口为查看这段时间消费失败的消息ID
```json
{
    "code": 0,
    "message": "",
    "data": [
        {
            "Score": 1597360839,//消息本来要在这一秒消费
            "Member": "15973617302072"//消息ID
        }
    ]
}
```

#### GET `/api/message/detail?msg_id=15771179759758&consumer=consumer_name`接口为查看消息详细信息
response:
```json
{
    "code": 0,
    "message": "",
    "data": {
        "id": 15971219465821,
        "cmd": "ddd:create:order",
        "timestamp": 1597121929,
        "params": "{\"user_id\":88}",
        "project": "sds",
        "bucket": "sds_bucket_5",
        "create_time": 1597121930,
        "request_id": ""
    }
}
```
#### POST `/api/message/solved`接口为将这些消息设置为已消费状态
request:
```json
{
    "consumer":"consumer_name",//消费者
    "msg_ids":[
        15973612939410//处理的消息ID
    ]
}
```
response:
```json
{
    "code": 0,
    "message": "",
    "data": {
        "15973612939410": true//消息ID和对应的处理结果
    }
}
```
### server端负责处理消息，到时间点时让消费者处理消息

## 使用
config目录是配置目录，具体看配置文件注释

假设已经编译好执行文件server和client

启动server `path/to/server ../config(相对路径，client的配置文件目录)`

启动client `path/to/client ../config(相对路径，client的配置文件目录)`

## 快速运行/编译脚本
go.sh为快速运行/编译脚本
```shell script
go.sh run|build server|client
```
dmq.sh为编译后运行脚本
```shell script
dmq.sh start|stop|restart client|server
```
