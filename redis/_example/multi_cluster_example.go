package main

import (
	"context"
	"fmt"
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
			Auth:             "qwe123",
			CodisClusterName: StarCodisClusterName,
			Opts: []redis.Option{
				redis.SetPoolSize(20),
				redis.SetConnectTimeout(200 * time.Millisecond),
				redis.SetReadTimeout(100 * time.Millisecond),
				redis.SetWriteTimeout(100 * time.Millisecond),
				redis.SetMaxConn(100),
				redis.SetAcquireConnMode(redis.AcquireConnModeUnblock),
				redis.EnableNodemgr(),
				redis.SetWorkerCycle(1),
				redis.SetHealthyThreshold(10),
				redis.SetMaxCooldownTime(60),
				redis.SetMinHealthyRatio(0.67),
			},
		},
		{
			Addrs:            []string{"127.0.0.1:6379"},
			Auth:             "qwe123",
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
	redisManager, _ = redis.NewManagerMap(configs)
}

func main() {
	starCodis, err := redisManager.GetManager(StarCodisClusterName)
	if err != nil {
		fmt.Println("get codis manager error", err)
	}

	res, err := starCodis.Set(context.TODO(), "accc", "to build nation by the people, for the people, of the people")
	if err != nil {
		fmt.Println("set error:", err)
	}

	res, err = starCodis.Get(context.TODO(), "accc")

	if err != nil {
		fmt.Println("get key error:", err)
	}

	fmt.Printf("get res from redis cluster %s:\n", StarCodisClusterName)
	fmt.Println(string(res.([]byte)))
}
