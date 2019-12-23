# dmq
golang 基于redis实现的消息队列，支持延时消息，使用了协程提高并发能力

## 分为client和server端

### client端负责接收消息，提供http接口

#### `/api/message/single`接口为提交一个消息
request:
```json
{
  "project":"sds", // 所属项目
  "cmd":"sds:create:order",// 消息执行的命令
  "timestamp":0,// 执行的时间点 时间戳 为0立即执行
  "params":"{\"order_id\":34256}",// 消息参数
  "bucket":"order:342543"// 命令bucket
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

#### `/api/message/batch`接口为批量提交消息
request:
```json
[
  {
    "project":"sds", // 所属项目
    "cmd":"sds:create:order",// 消息执行的命令
    "timestamp":0,// 执行的时间点 时间戳 为0立即执行
    "params":"{\"order_id\":34256}",// 消息参数
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

#### `/api/message/status?msg_id=15771179759758`接口为查看消息消费的状态
response:
```json
{
  "code": 0,
  "message": "",
  "data": {
    "dmq:consumer:10.190.40.90:8959/api/mytest/test/noticeuser": "3" // 消费者和对应的消费状态 1 未消费，2 正在消费，3 消费成功，4 消费失败
  }
}
```

### server端负责处理消息，到时间点时让消费者处理消息

## 使用
config目录是配置目录，具体看配置文件注释

假设已经编译好执行文件server和client

启动server `path/to/server ../config(相对路径，client的配置文件目录)`

启动client `path/to/client ../config(相对路径，client的配置文件目录)`