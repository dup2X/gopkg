// Package redis ...
package redis

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"testing"
	"time"

	dctx "github.com/dup2X/gopkg/context"
)

func assert(t *testing.T, exp bool) {
	if !exp {
		_, f, n, _ := runtime.Caller(1)
		println(f, n)
		t.FailNow()
	}
}

var addrs = []string{
	"127.0.0.1:6379",
}

var addrs2 = []string{
	"127.0.0.1:6379",
}

var ctx context.Context

func genConn() (*Manager, error) {
	ctx = context.Background()
	r, _ := http.NewRequest("GET", "dddd", nil)
	dctx.SetTrace(ctx, r)
	auth := "qwe123"
	mgr, err := NewManager(addrs2, auth, Prefix("test"), SetReadTimeout(time.Second*2),
		SetWriteTimeout(time.Second*2), SetPoolSize(8), SetClusterName("test"), SetKeepSilent(true),
		SetStatFunc(aaa))
	if err != nil {
		println(err.Error())
	}
	return mgr, err
}

// 这是需要传入的回调函数
func aaa(ctx context.Context, cmd string, cost time.Duration, err error) error {
	fmt.Printf("%s:::::::%v::::::::%v\n", cmd, cost, err)
	return nil
}

func TestManagerMap(t *testing.T) {
	auth := "qwe123"
	configs := []ManagerMapConfig{
		{
			Addrs:            []string{"127.0.0.1:6379"},
			Auth:             auth,
			CodisClusterName: "star",
			Opts: []Option{
				SetPoolSize(20),
				SetConnectTimeout(200 * time.Millisecond),
				SetReadTimeout(100 * time.Millisecond),
				SetWriteTimeout(100 * time.Millisecond),
				SetMaxConn(100),
				SetAcquireConnMode(AcquireConnModeUnblock),
			},
		},
	}
	redisManager, err := NewManagerMap(configs)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	want := "123"
	key := "one"
	c, err := redisManager.GetManager("star")
	assert(t, err == nil)
	defer c.Del(ctx, key)

	_, err = c.Set(ctx, key, want)
	assert(t, err == nil)
	existed, err := c.Exists(ctx, key)
	assert(t, err == nil && existed)

	_, err = c.Expire(ctx, key, 1024)
	assert(t, err == nil)

	got, err := c.GetString(ctx, key)
	assert(t, err == nil && got == want)

	_, err = c.Del(ctx, key)
	assert(t, err == nil)
	existed, err = c.Exists(ctx, key)
	assert(t, err == nil && existed == false)

	_, err = c.Incr(ctx, key)
	assert(t, err == nil)
	_, err = c.IncrBy(ctx, key, 1024)
	assert(t, err == nil)
	_, err = c.Decr(ctx, key)
	assert(t, err == nil)
	_, err = c.DecrBy(ctx, key, 1024)
	assert(t, err == nil)
}

func TestManagerMapWithNodemgr(t *testing.T) {
	auth := "qwe123"
	configs := []ManagerMapConfig{
		{
			Addrs:            []string{"127.0.0.1:6379", "127.0.0.1:6379"},
			Auth:             auth,
			CodisClusterName: "star",
			Opts: []Option{
				SetPoolSize(20),
				SetConnectTimeout(200 * time.Millisecond),
				SetReadTimeout(100 * time.Millisecond),
				SetWriteTimeout(100 * time.Millisecond),
				SetMaxConn(100),
				SetAcquireConnMode(AcquireConnModeUnblock),
				EnableNodemgr(),
				SetWorkerCycle(1),
				SetHealthyThreshold(10),
				SetMaxCooldownTime(60),
				SetMinHealthyRatio(0.67),
			},
		},
	}
	redisManager, err := NewManagerMap(configs)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	want := "123"
	key := "one"
	c, err := redisManager.GetManager("star")
	assert(t, err == nil)
	defer c.Del(ctx, key)

	_, err = c.Set(ctx, key, want)
	assert(t, err == nil)
	existed, err := c.Exists(ctx, key)
	assert(t, err == nil && existed)

	_, err = c.Expire(ctx, key, 1024)
	assert(t, err == nil)

	got, err := c.GetString(ctx, key)
	assert(t, err == nil && got == want)

	_, err = c.Del(ctx, key)
	assert(t, err == nil)
	existed, err = c.Exists(ctx, key)
	assert(t, err == nil && existed == false)

	_, err = c.Incr(ctx, key)
	assert(t, err == nil)
	_, err = c.IncrBy(ctx, key, 1024)
	assert(t, err == nil)
	_, err = c.Decr(ctx, key)
	assert(t, err == nil)
	_, err = c.DecrBy(ctx, key, 1024)
	assert(t, err == nil)
}

func TestAcquireConnMode(t *testing.T) {
	ctx = context.Background()
	r, _ := http.NewRequest("GET", "dddd", nil)
	dctx.SetTrace(ctx, r)
	auth := "qwe123"
	poolSize := int64(6)
	modes := []AcquireConnMode{AcquireConnModeBlock, AcquireConnModeUnblock, AcquireConnModeTimeout}
	for _, mode := range modes {
		c, err := NewManager(addrs, auth, Prefix("test"), SetReadTimeout(time.Second*2),
			SetWriteTimeout(time.Second*2), SetPoolSize(poolSize),
			SetWaitTimeout(time.Millisecond*500),
			SetConnectTimeout(time.Millisecond*50),
			SetAcquireConnMode(mode),
		)
		assert(t, err == nil)
		now := time.Now()
		for i := int64(0); i < poolSize; i++ {
			go c.mockBlock(time.Second * 1)
		}
		time.Sleep(time.Millisecond * 10)
		switch mode {
		case AcquireConnModeBlock:
			c.Get(ctx, "aa")
			assert(t, time.Now().Sub(now) >= time.Second)
		case AcquireConnModeUnblock:
			_, err = c.Get(ctx, "aa")
			assert(t, err != nil)
			assert(t, time.Now().Sub(now) < time.Second)
		case AcquireConnModeTimeout:
			_, err = c.Get(ctx, "aa")
			assert(t, err != nil)
			assert(t, time.Now().Sub(now) >= time.Millisecond*500)
		}
	}
}

func TestString(t *testing.T) {
	want := "123"
	key := "one"
	c, err := genConn()
	assert(t, err == nil)
	defer c.Del(ctx, key)

	_, err = c.Set(ctx, key, want)
	assert(t, err == nil)
	existed, err := c.Exists(ctx, key)
	assert(t, err == nil && existed)

	_, err = c.Expire(ctx, key, 1024)
	assert(t, err == nil)

	got, err := c.GetString(ctx, key)
	assert(t, err == nil && got == want)

	_, err = c.Del(ctx, key)
	assert(t, err == nil)
	existed, err = c.Exists(ctx, key)
	assert(t, err == nil && existed == false)

	_, err = c.Incr(ctx, key)
	assert(t, err == nil)
	_, err = c.IncrBy(ctx, key, 1024)
	assert(t, err == nil)
	_, err = c.Decr(ctx, key)
	assert(t, err == nil)
	_, err = c.DecrBy(ctx, key, 1024)
	assert(t, err == nil)

}

func TestMGet(t *testing.T) {
	c, err := genConn()
	assert(t, err == nil)

	kv := map[string]interface{}{
		"k1": 123,
		"k2": "v2",
	}
	defer func() {
		for k := range kv {
			c.Del(ctx, k)
		}
	}()

	_, err = c.MSet(ctx, kv)
	if err != nil {
		t.FailNow()
	}
	ret, err := c.MGet(ctx, []string{"k1", "k2"})
	if err != nil {
		t.FailNow()
	}
	if len(ret) != 2 || ret["k1"] != "123" || ret["k2"] != "v2" {
		t.FailNow()
	}
}

func TestHash(t *testing.T) {
	key := "test_hash"
	sub := []string{"s1", "s2"}
	subMap := make(map[string]interface{})
	for _, k := range sub {
		subMap[k] = "11"
	}
	c, err := genConn()
	assert(t, err == nil)
	defer c.Del(ctx, key)

	_, err = c.HMSet(ctx, key, subMap)
	assert(t, err == nil)
	ret, err := c.HMGet(ctx, key, sub)
	assert(t, err == nil)
	for _, k := range sub {
		assert(t, subMap[k] == ret[k])
	}

	sMap, err := c.HGetAll(ctx, key)
	assert(t, err == nil)
	for _, k := range sub {
		assert(t, subMap[k] == sMap[k])
	}

	for _, k := range sub {
		_, err = c.HSet(ctx, key, k, subMap[k])
		assert(t, err == nil)
		_, err = c.HGet(ctx, key, k)
		assert(t, err == nil)
	}

	keys, err := c.HKeys(ctx, key)
	assert(t, err == nil)
	assert(t, len(keys) == len(sub))

	length, err := c.HLen(ctx, key)
	assert(t, err == nil)
	assert(t, length == len(sub))

	for _, k := range sub {
		_, err = c.HDel(ctx, key, k)
		assert(t, err == nil)
		_, err = c.HIncrBy(ctx, key, k, 1024)
		assert(t, err == nil)
		ok, err := c.HExists(ctx, key, k)
		assert(t, err == nil)
		assert(t, ok == true)
	}
}

func TestList(t *testing.T) {
	key := "test_list"
	c, err := genConn()
	assert(t, err == nil)
	defer c.Del(ctx, key)

	_, err = c.LPush(ctx, key, "a")
	assert(t, err == nil)
	l, err := c.LLen(ctx, key)
	assert(t, err == nil)
	assert(t, 1 == l)
	vals, err := c.LRange(ctx, key, 0, 1)
	assert(t, err == nil)
	assert(t, 1 == len(vals))
	_, err = c.RPop(ctx, key)
	assert(t, err == nil)
	_, err = c.RPush(ctx, key, "a")
	assert(t, err == nil)
	l, err = c.LLen(ctx, key)
	assert(t, err == nil)
	assert(t, 1 == l)
	_, err = c.LPop(ctx, key)
	assert(t, err == nil)
	_, err = c.BLPop(ctx, key, 1)
	assert(t, err == nil)
	_, err = c.BRPop(ctx, key, 1)
	assert(t, err == nil)
	c.LPush(ctx, key, "ok")
	c.LPush(ctx, key, "ok")
	c.LPush(ctx, key, "ok1")
	_, err = c.LRem(ctx, key, 2, "ok")
	assert(t, err == nil)

}

func TestSetEx(t *testing.T) {
	key := "test_setex"
	c, err := genConn()
	assert(t, err == nil)
	defer c.Del(ctx, key)

	_, err = c.SetEx(ctx, key, 1, 2)
	assert(t, err == nil)
	time.Sleep(time.Second * 2)
	val, err := c.Get(ctx, key)
	assert(t, err == nil)
	assert(t, val == nil)
}

func TestSet(t *testing.T) {
	key := "test_set"
	c, err := genConn()
	assert(t, err == nil)
	defer c.Del(ctx, key)

	_, err = c.SAdd(ctx, key, "a")
	assert(t, err == nil)
	_, err = c.SAdd(ctx, key, "b")
	assert(t, err == nil)
	l, err := c.SCard(ctx, key)
	assert(t, err == nil)
	assert(t, 2 == l)
	members, err := c.SMembers(ctx, key)
	assert(t, err == nil)
	assert(t, 2 == len(members))
	ok, err := c.SIsMember(ctx, key, "a")
	assert(t, err == nil)
	assert(t, ok == true)
	ret, err := c.SUnion(ctx, []string{key})
	assert(t, err == nil)
	assert(t, 2 == len(ret))
	_, err = c.SRem(ctx, key, "a")
	assert(t, err == nil)
	_, err = String(c.SetNEx(ctx, "kLock", 344, "awdwer123123 20:49"))
	if err != nil {
		t.Fatalf("=====%s\n", err)
	}
}

func BenchmarkSet(b *testing.B) {
	auth := ""
	c, err := NewManager(addrs, auth, Prefix("test"), SetReadTimeout(time.Second*2),
		SetWriteTimeout(time.Second*2), SetPoolSize(256))
	if err != nil {
		b.FailNow()
	}

	want := "123"
	key := "one"

	for i := 0; i < b.N; i++ {
		_, err = c.Set(ctx, key, want)
		if err != nil {
			b.FailNow()
		}
	}

}
