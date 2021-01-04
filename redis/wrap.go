// Package redis defined redis_client
package redis

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
)

// String wrapper
func String(reply interface{}, err error) (string, error) {
	ret, err := redis.String(reply, err)
	if err == redis.ErrNil {
		err = &missedKeyErr{err}
	}
	return ret, err
}

// Strings wrapper
func Strings(reply interface{}, err error) ([]string, error) {
	// TODO check set stringSlice
	ret, err := redis.Strings(reply, err)
	if err == redis.ErrNil {
		err = &missedKeyErr{err}
	}
	return ret, err
}

// Int wrapper
func Int(reply interface{}, err error) (int, error) {
	ret, err := redis.Int(reply, err)
	if err == redis.ErrNil {
		err = &missedKeyErr{err}
	}
	return ret, err
}

// Ints wrapper
func Ints(reply interface{}, err error) ([]int, error) {
	ret, err := redis.Ints(reply, err)
	if err == redis.ErrNil {
		err = &missedKeyErr{err}
	}
	return ret, err
}

// Int64 wrapper
func Int64(reply interface{}, err error) (int64, error) {
	ret, err := redis.Int64(reply, err)
	if err == redis.ErrNil {
		err = &missedKeyErr{err}
	}
	return ret, err
}

// Bytes wrapper
func Bytes(reply interface{}, err error) ([]byte, error) {
	ret, err := redis.Bytes(reply, err)
	if err == redis.ErrNil {
		err = &missedKeyErr{err}
	}
	return ret, err
}

// ByteSlices wrapper
func ByteSlices(reply interface{}, err error) ([][]byte, error) {
	ret, err := redis.ByteSlices(reply, err)
	if err == redis.ErrNil {
		err = &missedKeyErr{err}
	}
	return ret, err
}

// Values wrapper
func Values(reply interface{}, err error) ([]interface{}, error) {
	ret, err := redis.Values(reply, err)
	if err == redis.ErrNil {
		err = &missedKeyErr{err}
	}
	return ret, err
}

func replyMap(reply interface{}, keys []string, err error) (map[string]string, error) {
	vals, err := redis.Strings(reply, err)
	if err == redis.ErrNil {
		err = &missedKeyErr{err}
	}
	if err != nil {
		return nil, err
	}
	size := len(keys)

	if size != len(vals) {
		return nil, fmt.Errorf("got %d valus, but %d wanted", len(vals), len(keys))
	}

	ret := make(map[string]string, size)
	for i := 0; i < size; i++ {
		ret[keys[i]] = vals[i]
	}
	return ret, nil
}
