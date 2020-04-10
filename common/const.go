package common

const (
	// redis key 前缀
	RedisKeyPrefix = "dmq"
	// 自增ID的key
	IdIncrKey = "message:id:generate:incr"

	MessageStatusWaiting = 1 //默认状态
	MessageStatusDoing   = 2 //正在消费
	MessageStatusDone    = 3 //消费完
	MessageStatusFailed  = 4 //消费失败
)
