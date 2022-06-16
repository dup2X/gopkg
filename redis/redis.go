// Package redis defined redis_client
package redis

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dup2X/gopkg/config"
	dctx "github.com/dup2X/gopkg/context"
	"github.com/dup2X/gopkg/discovery"
	"github.com/dup2X/gopkg/elapsed"

	"github.com/garyburd/redigo/redis"
)

const (
	defaultRetryTimes   = 2
	defaultPoolSize     = 16
	defaultWaitDuration = time.Millisecond * 10
)

// AcquireConnMode ...
type AcquireConnMode uint8

const (
	// AcquireConnModeUnblock ...
	AcquireConnModeUnblock AcquireConnMode = iota
	// AcquireConnModeTimeout ...
	AcquireConnModeTimeout
	// AcquireConnModeBlock ...
	AcquireConnModeBlock
)

// Conn wrapper
type Conn struct {
	redis.Conn
	addr string
}

// Manager redis client
type Manager struct {
	servers []string
	auth    string
	// connection pool
	conns         chan *Conn
	opt           *option
	localBalancer discovery.Balancer
	Connected     int // 初始化新建的可用连接数

	mu         *sync.Mutex
	activeConn int64
}

// NewManagerFromConfig ...
func NewManagerFromConfig(cfg config.Configer, sec string) (*Manager, error) {
	var opts []Option
	opts = append(opts, SetClusterName(sec))
	addrs, err := cfg.GetSetting(sec, "addrs")
	if err != nil {
		return nil, err
	}
	db, _ := cfg.GetIntSetting(sec, "db")
	if db > 0 {
		opts = append(opts, SetDB(int(db)))
	}
	servers := strings.Split(addrs, ",")
	auth, err := cfg.GetSetting(sec, "auth")
	if err != nil {
		return nil, err
	}
	disfEnable, _ := cfg.GetBoolSetting(sec, "disf_enable")
	if disfEnable {
		sn, err := cfg.GetSetting(sec, "service_name")
		if err != nil {
			return nil, err
		}
		opts = append(opts, EnableDisf())
		opts = append(opts, DisfServiceName(sn))
	}

	return NewManager(servers, auth, opts...)
}

// NewManager ...
func NewManager(addrs []string, auth string, opts ...Option) (*Manager, error) {
	var err error
	opt := &option{}
	for _, o := range opts {
		o(opt)
	}
	if opt.poolSize == 0 {
		opt.poolSize = defaultPoolSize
	}
	if opt.maxTryTimes == 0 {
		opt.maxTryTimes = defaultRetryTimes
	}
	if opt.maxConn < opt.poolSize {
		opt.maxConn = opt.poolSize
	}
	if opt.mode == AcquireConnModeUnblock {
		opt.waitTimeout = 0
	} else if opt.mode == AcquireConnModeTimeout && opt.waitTimeout == 0 {
		opt.waitTimeout = defaultWaitDuration
	}

	mgr := &Manager{
		servers: addrs,
		auth:    auth,
		opt:     opt,
		mu:      new(sync.Mutex),
	}

	mgr.localBalancer, _ = discovery.NewBalancer(
		discovery.LOCALTYPE,
		"",
		addrs,
	)

	cnt, err := mgr.initPool()
	mgr.Connected = cnt
	if mgr.opt.keepSilent && mgr.Connected > 0 {
		return mgr, nil
	}
	return mgr, err
}

// init connection pool
func (m *Manager) initPool() (usable int, err error) {
	m.conns = make(chan *Conn, m.opt.poolSize)
	for i := int64(0); i < m.opt.poolSize; i++ {
		conn, err := m.newConn()
		if err != nil {
			if m.opt.keepSilent {
				continue
			}
			return usable, err
		}
		_, err = conn.Do(commandPing)
		if err != nil {
			m.voteUnhealthy(conn.addr)
			if m.opt.keepSilent {
				continue
			}
			return usable, err
		}
		m.voteHealthy(conn.addr)
		m.putConn(conn)
		usable++
	}
	return usable, nil
}

func (m *Manager) newConn() (*Conn, error) {
	var (
		err  error
		addr string
	)

	if addr == "" || err != nil {
		if addr == "" || err != nil {
			addr, err = m.localBalancer.Get()
		}

		if err != nil {
			return nil, err
		}
	}

	var opts []redis.DialOption
	if m.opt.db > 0 {
		opts = append(opts, redis.DialDatabase(m.opt.db))
	}
	if m.opt.readTimeout > 0 {
		opts = append(opts, redis.DialReadTimeout(m.opt.readTimeout))
	}
	if m.opt.writeTimeout > 0 {
		opts = append(opts, redis.DialWriteTimeout(m.opt.writeTimeout))
	}
	if m.opt.connTimeout > 0 {
		opts = append(opts, redis.DialConnectTimeout(m.opt.connTimeout))
	}
	if m.auth != "" {
		opts = append(opts, redis.DialPassword(m.auth))
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.activeConn >= m.opt.maxConn {
		return nil, fmt.Errorf("too much conns")
	}
	c, err := redis.Dial("tcp4", addr, opts...)
	if err != nil {
		m.voteUnhealthy(addr)
		return nil, err
	}
	conn := &Conn{c, addr}
	/*
		if m.auth != "" {
			_, err = conn.Do(commandAuth, m.auth)
			if err != nil {
				m.voteUnhealthy(addr)
				conn.Close()
				return nil, err
			}
		}
	*/
	m.voteHealthy(addr)
	m.activeConn++
	return conn, err
}

func (m *Manager) voteHealthy(addr string) error {
	return nil
}

func (m *Manager) voteUnhealthy(addr string) error {
	return nil
}

func (m *Manager) getConn() (*Conn, error) {
	switch m.opt.mode {
	case AcquireConnModeTimeout:
		return m.getConnTimeout()
	case AcquireConnModeBlock:
		return <-m.conns, nil
	case AcquireConnModeUnblock:
		return m.getConnUnblock()
	default:
		return nil, fmt.Errorf("unsupported acquire conn mode: %d", m.opt.mode)
	}
}

func (m *Manager) getConnTimeout() (*Conn, error) {
	select {
	case conn := <-m.conns:
		return conn, nil
	case <-time.After(m.opt.waitTimeout):
		return nil, ErrAcquiredConnTimeout
	}
}

func (m *Manager) getConnUnblock() (*Conn, error) {
	select {
	case conn := <-m.conns:
		return conn, nil
	default:
		return nil, ErrEmptyConnPool
	}
}

func (m *Manager) putConn(conn *Conn) {
	select {
	case m.conns <- conn:
	default:
		m.mu.Lock()
		conn.Close()
		m.activeConn--
		m.mu.Unlock()
	}
}

func (m *Manager) do(ctx context.Context, action func(*Conn) (interface{}, error), cmd string, arg ...interface{}) (interface{}, error) {
	if m.opt.slaFuse && !dctx.CheckSLA(ctx) {
		return nil, ErrSLATimeout
	}
	et := elapsed.New()
	et.Start()
	ret, err := m.redialDo(ctx, action)
	cost := et.Stop()

	if m.opt.statFunc != nil {
		cmdStr := cmd
		for _, v := range arg {
			cmdStr += " " + fmt.Sprint(v)
		}
		m.opt.statFunc(ctx, cmdStr, cost, err)
	}

	dctx.AddRedisElapsed(ctx, cost)
	return ret, err
}

func (m *Manager) redialDo(ctx context.Context, action func(conn *Conn) (interface{}, error)) (reply interface{}, err error) {
	conn, err := m.getConn()
	tried := int64(0)
	if err != nil {
		goto retry
	}

start:
	tried++
	if tried > m.opt.maxTryTimes {
		m.putConn(conn)
		return
	}
	reply, err = action(conn)
	if err != nil {
		m.voteUnhealthy(conn.addr)
		goto retry
	}
	m.voteHealthy(conn.addr)
	m.putConn(conn)
	return

retry:
	if conn != nil {
		m.mu.Lock()
		conn.Close()
		m.activeConn--
		m.mu.Unlock()
	}
	var (
		newConn *Conn
		newErr  error
	)
	if newConn, newErr = m.newConn(); newErr != nil {
		err = newErr
		return
	}
	conn = newConn
	goto start
}

// Set command
func (m *Manager) Set(ctx context.Context, key string, val interface{}) (reply interface{}, err error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandSet, key, val)
	}
	return m.do(ctx, action, commandSet, key, val)
}

// SetEx command
func (m *Manager) SetEx(ctx context.Context, key string, expireTime int, val interface{}) (reply interface{}, err error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandSetEx, key, expireTime, val)
	}
	return m.do(ctx, action, commandSetEx, key, expireTime, val)
}

// SetNEx setNX与Expire的合并 需要redis版本大于2.6.12
func (m *Manager) SetNEx(ctx context.Context, key string, expireTime int, val interface{}) (reply interface{}, err error) {
	strVal := fmt.Sprint(val)
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandSet, key, strVal, "EX", expireTime, "NX")
	}
	return m.do(ctx, action, commandSet, key, strVal, "EX", expireTime, "NX")
}

// Get command
func (m *Manager) Get(ctx context.Context, key string) (reply interface{}, err error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandGet, key)
	}
	return m.do(ctx, action, commandGet, key)
}

// Incr command
func (m *Manager) Incr(ctx context.Context, key string) (reply interface{}, err error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandIncr, key)
	}
	return m.do(ctx, action, commandIncr, key)
}

// IncrBy command
func (m *Manager) IncrBy(ctx context.Context, key string, delt int) (reply interface{}, err error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandIncrBy, key, delt)
	}
	return m.do(ctx, action, commandIncrBy, key, delt)
}

// Decr command
func (m *Manager) Decr(ctx context.Context, key string) (reply interface{}, err error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandDecr, key)
	}
	return m.do(ctx, action, commandDecr, key)
}

// DecrBy command
func (m *Manager) DecrBy(ctx context.Context, key string, delt int) (reply interface{}, err error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandDecrBy, key, delt)
	}
	return m.do(ctx, action, commandDecrBy, key, delt)
}

// MSet command
func (m *Manager) MSet(ctx context.Context, kv map[string]interface{}) (reply interface{}, err error) {
	var kvList []interface{}
	for k, v := range kv {
		kvList = append(kvList, k, v)
	}
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandMSet, kvList...)
	}
	return m.do(ctx, action, commandMSet, kvList...)
}

// MGet command
func (m *Manager) MGet(ctx context.Context, keys []string) (map[string]string, error) {
	var keyList []interface{}
	for _, k := range keys {
		keyList = append(keyList, k)
	}
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandMGet, keyList...)
	}
	reply, err := m.do(ctx, action, commandMGet, keyList...)
	return replyMap(reply, keys, err)
}

// Exists command
func (m *Manager) Exists(ctx context.Context, key string) (bool, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandExists, key)
	}
	return redis.Bool(m.do(ctx, action, commandExists, key))
}

// Expire command
func (m *Manager) Expire(ctx context.Context, key string, ttl int64) (interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandExpire, key, ttl)
	}
	return m.do(ctx, action, commandExpire, key, ttl)
}

// Del command
func (m *Manager) Del(ctx context.Context, key string) (interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandDel, key)
	}
	return m.do(ctx, action, commandDel, key)
}

// HExists command
func (m *Manager) HExists(ctx context.Context, key, sub string) (bool, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandHExists, key, sub)
	}
	return redis.Bool(m.do(ctx, action, commandHExists, key, sub))
}

// HGet command
func (m *Manager) HGet(ctx context.Context, key, sub string) (reply interface{}, err error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandHGet, key, sub)
	}
	return m.do(ctx, action, commandHGet, key, sub)
}

// HGetAll 线上应该谨慎(禁止)使用，subKey数目过多时会阻塞请求,影响性能
func (m *Manager) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandHGetAll, key)
	}
	return redis.StringMap(m.do(ctx, action, commandHGetAll, key))
}

// HSet command
func (m *Manager) HSet(ctx context.Context, key, sub string, val interface{}) (reply interface{}, err error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandHSet, key, sub, val)
	}
	return m.do(ctx, action, commandHSet, key, sub, val)
}

// HIncrBy command
func (m *Manager) HIncrBy(ctx context.Context, key, sub string, delt int) (reply interface{}, err error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandHIncrBy, key, sub, delt)
	}
	return m.do(ctx, action, commandHIncrBy, key, sub, delt)
}

// HMSet command
func (m *Manager) HMSet(ctx context.Context, key string, subKV map[string]interface{}) (interface{}, error) {
	var kvList []interface{}
	kvList = append(kvList, key)
	for k, v := range subKV {
		kvList = append(kvList, k, v)
	}
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandHMSet, kvList...)
	}
	return m.do(ctx, action, commandHMSet, kvList...)
}

// HMGet command
func (m *Manager) HMGet(ctx context.Context, key string, subKeys []string) (map[string]string, error) {
	var args []interface{}
	args = append(args, key)
	for _, k := range subKeys {
		args = append(args, k)
	}
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandHMGet, args...)
	}
	reply, err := m.do(ctx, action, commandHMGet, args...)
	return replyMap(reply, subKeys, err)
}

// HLen command
func (m *Manager) HLen(ctx context.Context, key string) (int, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandHLen, key)
	}
	return redis.Int(m.do(ctx, action, commandHLen, key))
}

// HDel command
func (m *Manager) HDel(ctx context.Context, key, sub string) (interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandHDel, key, sub)
	}
	return m.do(ctx, action, commandHDel, key, sub)
}

// HKeys command
func (m *Manager) HKeys(ctx context.Context, key string) ([]string, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandHKeys, key)
	}
	return redis.Strings(m.do(ctx, action, commandHKeys, key))
}

// LPush command
func (m *Manager) LPush(ctx context.Context, key string, val interface{}) (interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandLPush, key, val)
	}
	return m.do(ctx, action, commandLPush, key, val)
}

// RPush command
func (m *Manager) RPush(ctx context.Context, key string, val interface{}) (interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandRPush, key, val)
	}
	return m.do(ctx, action, commandRPush, key, val)
}

// LPop command
func (m *Manager) LPop(ctx context.Context, key string) (interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandLPop, key)
	}
	return m.do(ctx, action, commandLPop, key)
}

// RPop command
func (m *Manager) RPop(ctx context.Context, key string) (interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandRPop, key)
	}
	return m.do(ctx, action, commandRPop, key)
}

// BLPop command
func (m *Manager) BLPop(ctx context.Context, key string, secTimeout int64) (interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandBLPop, key, secTimeout)
	}
	return m.do(ctx, action, commandBLPop, key, secTimeout)
}

// BRPop command
func (m *Manager) BRPop(ctx context.Context, key string, secTimeout int64) (interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandBRPop, key, secTimeout)
	}
	return m.do(ctx, action, commandBRPop, key, secTimeout)
}

// LLen command
func (m *Manager) LLen(ctx context.Context, key string) (int, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandLLen, key)
	}
	return redis.Int(m.do(ctx, action, commandLLen, key))
}

// LRange command
func (m *Manager) LRange(ctx context.Context, key string, start, end int) ([]interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandLRange, key, start, end)
	}
	return redis.Values(m.do(ctx, action, commandLRange, key, start, end))
}

// LRem command
func (m *Manager) LRem(ctx context.Context, key string, count int, val interface{}) (int64, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandLRem, key, count, val)
	}
	return redis.Int64(m.do(ctx, action, commandLRem, key, count, val))
}

// SAdd command
func (m *Manager) SAdd(ctx context.Context, key string, member interface{}) (interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandSAdd, key, member)
	}
	return m.do(ctx, action, commandSAdd, key, member)
}

// SCard command
func (m *Manager) SCard(ctx context.Context, key string) (int, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandSCard, key)
	}
	return redis.Int(m.do(ctx, action, commandSCard, key))
}

// SIsMember command
func (m *Manager) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandSIsMember, key, member)
	}
	return redis.Bool(m.do(ctx, action, commandSIsMember, key, member))
}

// SMembers command
func (m *Manager) SMembers(ctx context.Context, key string) ([]interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandSMembers, key)
	}
	return redis.Values(m.do(ctx, action, commandSMembers, key))
}

// SRem command
func (m *Manager) SRem(ctx context.Context, key string, member interface{}) (interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandSRem, key, member)
	}
	return m.do(ctx, action, commandSRem, key, member)
}

// ZAdd command
func (m *Manager) ZAdd(ctx context.Context, key string, score float64, member interface{}) (interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandZAdd, key, score, member)
	}
	return m.do(ctx, action, commandZAdd, key, score, member)
}

// ZRangeByScore command
func (m *Manager) ZRangeByScore(ctx context.Context, key string, min, max float64) (interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandZRangeByScore, key, min, max)
	}
	return m.do(ctx, action, commandZRangeByScore, key, min, max)
}

// ZRemRangeByScore command
func (m *Manager) ZRemRangeByScore(ctx context.Context, key string, min, max float64) (interface{}, error) {
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandZRemRangeByScore, key, min, max)
	}
	return m.do(ctx, action, commandZRemRangeByScore, key, min, max)
}

// SUnion command
func (m *Manager) SUnion(ctx context.Context, sets []string) ([]interface{}, error) {
	var list []interface{}
	for _, s := range sets {
		list = append(list, s)
	}
	action := func(conn *Conn) (interface{}, error) {
		return conn.Do(commandSUnion, list...)
	}
	return redis.Values(m.do(ctx, action, commandSUnion, list...))
}

// GetString wrapper get and string
func (m *Manager) GetString(ctx context.Context, key string) (reply string, err error) {
	return redis.String(m.Get(ctx, key))
}

func (m *Manager) mockBlock(sec time.Duration) {
	conn, err := m.getConn()
	if err != nil {
		println(err.Error())
		return
	}
	<-time.NewTimer(sec).C
	m.putConn(conn)
}

// MockBlock ...
func (m *Manager) MockBlock(dur time.Duration) {
	m.mockBlock(dur)
}
