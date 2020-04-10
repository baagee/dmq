package common

const (
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

	// 保存消息的lua脚本
	SaveMessageLuaScript = `local pointGroupName = KEYS[1]
local pointName = KEYS[2]
local bucketName = KEYS[3]
local messageListName = KEYS[4]
local messageStatusHashKey = KEYS[5]
local id2hashDayKey = KEYS[6]
local id2hashIdFieldKey = KEYS[7]
local messageDetailKey = KEYS[8]
local pointScore = tonumber(ARGV[1])
local msgHash = ARGV[2]
local expireTime = tonumber(ARGV[3])
local msgStr = ARGV[4]

-- 1 保存point
redis.call('ZADD', pointGroupName, pointScore, pointName)

-- 2 增加时间点的bucket
redis.call('SADD', pointName, bucketName)

-- 3 将任务hash放入当前时刻当前bucket的任务列表
redis.call('LPUSH', messageListName, msgHash)

-- 4 将任务状态保存
if (tonumber(string.sub(_VERSION,5))>=5.2) then
    redis.call('HMSET', messageStatusHashKey, table.unpack(ARGV, 5))
else
    redis.call('HMSET', messageStatusHashKey, unpack(ARGV, 5))
end

redis.call('EXPIRE', messageStatusHashKey, 2*expireTime)

-- 5 增加 project Y-m-d-H 按天纬度记录ID和hash 过期时间 expireTime id=>hash的关系 后面可以根据hash获取消息详情
redis.call('HSET', id2hashDayKey, id2hashIdFieldKey, msgHash)
redis.call('EXPIRE', id2hashDayKey, expireTime)

-- 6 增加消息标记 便于去重 通过hash可以查看m详情 hash=>m 保存消息全部信息 key=hash value=json_encode(m)
redis.call('SETEX', messageDetailKey, expireTime, msgStr)

return 1`
)
