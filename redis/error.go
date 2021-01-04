// Package redis defined redis_client
package redis

import (
	"errors"

	"github.com/garyburd/redigo/redis"
)

var (
	// ErrEmptyConnPool empty connection pool
	ErrEmptyConnPool = errors.New("empty connection pool")
	// ErrAcquiredConnTimeout acquire connection timed out
	ErrAcquiredConnTimeout = errors.New("acquire connection timed out")
	// ErrNilValue redis nil value
	ErrNilValue = redis.ErrNil
	// ErrNoManagerAvailable 表示在全局的 manager map 中找不到该 Manager 的注册记录
	ErrNoManagerAvailable = errors.New("no available manager in manager map, check your cluster name?")
	// ErrSLATimeout 已经超时直接熔断redis操作
	ErrSLATimeout = errors.New("sla timeout, interrupted redis action")
)

// Error 内部封装，为了区分出 Get 操作时 redis 返回值是否为 nil
type Error interface {
	error
	MissedKey() bool
}

const notFoundKey = "key is not found"

type missedKeyErr struct {
	err error
}

func (mke *missedKeyErr) Error() string {
	return notFoundKey
}

func (mke *missedKeyErr) MissedKey() bool {
	return true
}
