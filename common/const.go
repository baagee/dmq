package common

const (
	// redis key 前缀
	RedisKeyPrefix = "dmq"
	// 自增ID的key
	IdIncrKey = "message:id:generate:incr"
	// 获取时间点的lua脚本
	GetTimePointLuaScript = `local key = KEYS[1]
-- max score
local _end= ARGV[1]

local res=redis.call('ZRANGEBYSCORE', key, 0, _end, 'WITHSCORES', 'LIMIT', 0, 1)
if (#res == 0) then
	-- empty return 0
    return 0
else
    if (redis.call('ZREM', key, res[1]) == 1) then
    -- rem success return res
        return res
    else
    -- rem failed return false
        return false
    end
end`

	MessageStatusWaiting = 1 //默认状态
	MessageStatusDoing   = 2 //正在消费
	MessageStatusDone    = 3 //消费完
	MessageStatusFailed  = 4 //消费失败
)
