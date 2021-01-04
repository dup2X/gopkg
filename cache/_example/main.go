package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"sync/atomic"
	"time"

	"github.com/dup2X/gopkg/cache"
)

func main() {
	f, err := os.Create("example.prof")
	if err != nil {
		println(err.Error())
		return
	}
	pprof.StartCPUProfile(f)
	go func() {
		time.Sleep(time.Second * 20)
		pprof.StopCPUProfile()
	}()

	testPool()
}

func testPool() {
	pool := cache.NewMemPool(10240)
	sizes := []int{256, 512, 1024}
	var qps int64

	for k := 0; k < 10240; k++ {
		go func() {
			for i := 0; i < 1e8; i++ {
				data := pool.Get(sizes[i%3])
				pool.Put(data)
				atomic.AddInt64(&qps, 1)
			}
		}()
	}
	tk := time.NewTicker(time.Second * 1)
	for range tk.C {
		println("####", time.Now().Unix())
		println(atomic.SwapInt64(&qps, 0))
	}
}

func testCache() {
	var max int = 1e6
	var keys = make([][]byte, max)
	var val = []byte{'a', '1'}
	c := cache.NewLocal()
	for i := 0; i < max; i++ {
		keys[i] = []byte(fmt.Sprintf("teststststststststst123123,mh1kj2h3kjh12kj3-%05d", i))
		c.Set(keys[i], val)
	}
	var qps int64
	go func() {
		tk := time.NewTicker(time.Second)
		for range tk.C {
			println(atomic.SwapInt64(&qps, 0))
		}
	}()
	go func() {
		for i := 0; i < max; i++ {
			keys[i] = []byte(fmt.Sprintf("teststststststststst123123,mh1kj2h3kjh12kj3-%05d", i))
			c.Set(keys[i], val)
			time.Sleep(time.Millisecond)
		}
	}()
	for i := 0; i < 1e12; i++ {
		c.Get(keys[i%max])
		atomic.AddInt64(&qps, 1)
	}
}
