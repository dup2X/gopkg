package main

import (
	"sync/atomic"
	"time"

	dredis "github.com/dup2X/gopkg/redis"

	"github.com/garyburd/redigo/redis"
)

var qps uint64

func newClient() {
	var newFunc = func() (redis.Conn, error) {
		tw := time.Second * 2
		return redis.DialTimeout("tcp4", "127.0.0.1:6379", tw, tw, tw)
	}
	pool := redis.NewPool(newFunc, 256)

	for index := 0; index < 256; index++ {
		go func() {
			for {
				conn := pool.Get()
				if conn != nil {
					conn.Do("SET", "one", 123)
					atomic.AddUint64(&qps, 1)
					conn.Close()
				}
			}
		}()
	}
}

func test() {
	addrs := []string{"127.0.0.1:6379"}
	auth := ""

	c, err := dredis.NewManager(addrs, auth, dredis.Prefix("test"),
		dredis.SetReadTimeout(time.Second*2),
		dredis.SetWaitTimeout(time.Millisecond*30),
		dredis.SetAcquireConnMode(dredis.AcquireConnModeUnblock),
		dredis.SetWriteTimeout(time.Second*2), dredis.SetPoolSize(256))
	if err != nil {
		println(err.Error())
		return
	}
	for index := 0; index < 256; index++ {
		go func() {
			c.MockBlock(time.Second * 20)
		}()
		time.Sleep(time.Millisecond * 20)
		go func() {
			for {
				c.Set("one", 123)
				atomic.AddUint64(&qps, 1)
			}
		}()
	}
}

func main() {
	//	go newClient()
	go test()
	tk := time.NewTicker(time.Second)
	for range tk.C {
		println(atomic.SwapUint64(&qps, 0))
	}
}
