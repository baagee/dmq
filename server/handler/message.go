package handler

import (
	"github.com/baagee/dmq/common"
	"github.com/go-redis/redis"
)

func GetTimePoint(project string) (redis.Z, error) {
	zRes, err := common.RedisCli.ZRangeWithScores(common.GetPointGroup(project), 0, 0).Result()
	if err != nil {
		return redis.Z{}, common.ThrowNotice(common.ErrorCodeFoundPointFailed, err)
	}
	if len(zRes) == 1 {
		return zRes[0], nil
	}
	return redis.Z{}, nil
}

func RemoveTimePoint(project string, point string) error {
	_, err := common.RedisCli.ZRem(common.GetPointGroup(project), point).Result()
	if err != nil {
		return common.ThrowNotice(common.ErrorCodeRemovePointFailed, err)
	}
	return nil
}
