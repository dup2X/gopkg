package main

import (
	"context"
	"time"

	"github.com/dup2X/gopkg/redis"
)

var redisManager *redis.ManagerMap

// codis 集群业务名枚举
const (
	StarCodisClusterName       = "star"
	FinishRateCodisClusterName = "finishrate"
)

func init() {
	configs := []redis.ManagerMapConfig{
		{
			Addrs:            []string{"127.0.0.1:6379"},
			Auth:             "",
			CodisClusterName: StarCodisClusterName,
			Opts: []redis.Option{
				redis.SetPoolSize(20),
				redis.SetConnectTimeout(200 * time.Millisecond),
				redis.SetReadTimeout(100 * time.Millisecond),
				redis.SetWriteTimeout(100 * time.Millisecond),
				redis.SetMaxConn(100),
				redis.SetAcquireConnMode(redis.AcquireConnModeUnblock),
			},
		},
		{
			Addrs:            []string{"127.0.0.1:6379"},
			Auth:             "",
			CodisClusterName: StarCodisClusterName,
			Opts: []redis.Option{
				redis.SetPoolSize(20),
				redis.SetConnectTimeout(200 * time.Millisecond),
				redis.SetReadTimeout(100 * time.Millisecond),
				redis.SetWriteTimeout(100 * time.Millisecond),
				redis.SetMaxConn(100),
				redis.SetAcquireConnMode(redis.AcquireConnModeUnblock),
			},
		},
	}
	redisManager = redis.NewManagerMap(configs)
}

func main() {
	for i := 0; i < 2000; i++ {
		go func() {
			for {
				starCodis, _ := redisManager.GetManager(StarCodisClusterName)
				starCodis.Set(context.TODO(), "abcdefg", "big brother is watching you")
			}
		}()
	}

	time.Sleep(time.Hour)
}
