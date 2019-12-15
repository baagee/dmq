package util

import (
	"fmt"
	"math/rand"
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
