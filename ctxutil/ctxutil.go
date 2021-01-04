// Package ctxutil provides some helper functions which makes the
// context.Context object seemed to be an typed struct
// You don't need to know the keys just Set and Get
//
//
// Attention:
// Before calling IncDBTime or IncRedisTime to accumulate the overall time
// you need call SetDBTime first.
package ctxutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type keytype int

const (
	_ keytype = iota
	traceIDKey
	spanIDKey
	hintCodeKey
	hintContentKey
	langKey

	redisTimeKey
	dbTimeKey
	tokenKey

	callerKey
	degradeKey

	httpRequestKey
	requestInTsKey    // 请求进来的时间戳
	processTimeoutKey //上游超时要求
	cookiesKey
)

var (
	// ErrNotExist specifies that the data you want to retrive is missed
	ErrNotExist = errors.New("key is not existed in the context")
	// ErrDurationNotSetted reports that duration must be setted before you use it to do accumulation
	ErrDurationNotSetted = errors.New("duration must be setted at the beginning of the goroutine")
	//ErrWrongType ...
	ErrWrongType = errors.New("wrong type")
)

// helper
func getString(ctx context.Context, key keytype) (string, error) {
	t := ctx.Value(key)
	if t == nil {
		return "", ErrNotExist
	}
	return t.(string), nil
}

func getInt64(ctx context.Context, key keytype) (int64, error) {
	t := ctx.Value(key)
	if t == nil {
		return 0, ErrNotExist
	}
	return t.(int64), nil
}

func getTimeDuration(ctx context.Context, key keytype) (time.Duration, error) {
	v := ctx.Value(key)
	if nil == v {
		return time.Duration(0), ErrNotExist
	}
	return time.Duration(atomic.LoadInt64(v.(*int64))), nil
}

func incTimeDuration(ctx context.Context, key keytype, delta time.Duration) (context.Context, error) {
	v := ctx.Value(key)
	if v == nil {
		return ctx, ErrDurationNotSetted
	}
	vd := v.(*int64)
	atomic.AddInt64(vd, delta.Nanoseconds())
	return ctx, nil
}

func newI64Pointer() *int64 {
	a := new(int64)
	*a = 0
	return a
}

// SetTraceID sets traceID into context
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID gets traceID from context,
// if traceID not exists, ErrNotExist will be returned
func GetTraceID(ctx context.Context) (string, error) {
	return getString(ctx, traceIDKey)
}

// SetSpanID sets spanID into context
func SetSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, spanIDKey, spanID)
}

// GetSpanID gets spanID from context,
// if spanID not exists, ErrNotExist will be returned
func GetSpanID(ctx context.Context) (string, error) {
	return getString(ctx, spanIDKey)
}

func SetDBTime(ctx context.Context) context.Context {
	return context.WithValue(ctx, dbTimeKey, newI64Pointer())
}

// IncDBTime accumulates the time for operating database
// you must call SetDBTime before use this function
func IncDBTime(ctx context.Context, delta time.Duration) (context.Context, error) {
	return incTimeDuration(ctx, dbTimeKey, delta)
}

// GetDBTime returns the accumulated time for operating database
func GetDBTime(ctx context.Context) (time.Duration, error) {
	return getTimeDuration(ctx, dbTimeKey)
}

// SetRedisTime initialize the time counter for redis
func SetRedisTime(ctx context.Context) context.Context {
	return context.WithValue(ctx, redisTimeKey, newI64Pointer())
}

// IncRedisTime accumulates the time for operating redis
// you must call SetRedisTime before use this function
func IncRedisTime(ctx context.Context, delta time.Duration) (context.Context, error) {
	return incTimeDuration(ctx, redisTimeKey, delta)
}

// GetRedisTime returns the accumulated time for operating redis
func GetRedisTime(ctx context.Context) (time.Duration, error) {
	return getTimeDuration(ctx, redisTimeKey)
}

// SetHintCode sets the hintCode into context
func SetHintCode(ctx context.Context, hintCode int64) context.Context {
	return context.WithValue(ctx, hintCodeKey, hintCode)
}

// GetHintCode gets hintCode from context,
// if hintCode not exists, ErrNotExist will be returned
func GetHintCode(ctx context.Context) (int64, error) {
	return getInt64(ctx, hintCodeKey)
}

// SetHintContent sets hintContent into context
func SetHintContent(ctx context.Context, hintContent string) context.Context {
	return context.WithValue(ctx, hintContentKey, hintContent)
}

// GetHintContent gets hintContent from context,
// if hintContent not exists, ErrNotExist will be returned
func GetHintContent(ctx context.Context) (string, error) {
	return getString(ctx, hintContentKey)
}

// SetLang sets language params into context
func SetLang(ctx context.Context, language string) context.Context {
	return context.WithValue(ctx, langKey, language)
}

// GetLang gets language from context
// if langguage not exists, ErrNotExist will be returned
func GetLang(ctx context.Context) (string, error) {
	return getString(ctx, langKey)
}

// SetToken sets token into context
func SetToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

// GetToken ...
func GetToken(ctx context.Context) (string, error) {
	return getString(ctx, tokenKey)
}

// SetCaller ...
func SetCaller(ctx context.Context, caller string) context.Context {
	return context.WithValue(ctx, callerKey, caller)
}

// GetCaller ...
func GetCaller(ctx context.Context) string {
	v := ctx.Value(callerKey)
	if v == nil {
		return ""
	}
	if val, ok := v.(string); ok {
		return val
	}
	return ""
}

// SetDegrade ...
func SetDegrade(ctx context.Context, sed int) context.Context {
	return context.WithValue(ctx, degradeKey, sed)
}

// GetDegrade ...
func GetDegrade(ctx context.Context) int {
	v := ctx.Value(degradeKey)
	if v == nil {
		return 0
	}
	if val, ok := v.(int); ok {
		return val
	}
	return 0
}

//SetHTTPRequest ...
func SetHTTPRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, httpRequestKey, r)
}

//GetHTTPRequest ...
func GetHTTPRequest(ctx context.Context) (*http.Request, error) {
	v := ctx.Value(httpRequestKey)
	if v == nil {
		return nil, ErrNotExist
	}
	val, ok := v.(*http.Request)
	if !ok {
		return nil, ErrWrongType
	}
	return val, nil
}

// SetRequestInTs 请求的时间戳  毫秒位
func SetRequestInTs(ctx context.Context, ts int64) context.Context {
	return context.WithValue(ctx, requestInTsKey, ts)
}

// SetRequestTimeout 上游超时要求 毫秒位
func SetRequestTimeout(ctx context.Context, to int64) context.Context {
	return context.WithValue(ctx, processTimeoutKey, to)
}

// GetRequestInTs 获取请求的时间戳  毫秒位
func GetRequestInTs(ctx context.Context) (int64, error) {
	v := ctx.Value(requestInTsKey)
	if v == nil {
		return 0, ErrNotExist
	}
	val, ok := v.(int64)
	if !ok {
		return 0, ErrWrongType
	}
	return val, nil
}

// GetRequestTimeout 上游超时要求 毫秒位
func GetRequestTimeout(ctx context.Context) (int64, error) {
	v := ctx.Value(processTimeoutKey)
	if v == nil {
		return 0, ErrNotExist
	}
	val, ok := v.(int64)
	if !ok {
		return 0, ErrWrongType
	}
	return val, nil
}

func SetCookies(ctx context.Context, cks []*http.Cookie) context.Context {
	return context.WithValue(ctx, cookiesKey, cks)
}

func GetCookies(ctx context.Context) ([]*http.Cookie, error) {
	v := ctx.Value(cookiesKey)
	if v == nil {
		return nil, ErrNotExist
	}
	val, ok := v.([]*http.Cookie)
	if !ok {
		return nil, ErrWrongType
	}
	return val, nil
}

var builtInKeys = []keytype{
	traceIDKey,
	spanIDKey,
	redisTimeKey,
	dbTimeKey,
	hintCodeKey,
	hintContentKey,
	langKey,
	tokenKey,
	requestInTsKey,
	processTimeoutKey,
}

var keyName = map[keytype]string{
	traceIDKey:        "traceid",
	spanIDKey:         "spanid",
	redisTimeKey:      "redisTime",
	dbTimeKey:         "dbTime",
	hintCodeKey:       "hintCode",
	hintContentKey:    "hintContent",
	langKey:           "lang",
	tokenKey:          "token",
	requestInTsKey:    "request_in_ts",
	processTimeoutKey: "process_timeout",
}

func shouldConvert2Duration(k keytype) bool {
	return k == redisTimeKey || k == dbTimeKey
}

func getVal(ctx context.Context, key keytype) (interface{}, error) {
	if ctx == nil {
		return nil, ErrNotExist
	}
	v := ctx.Value(key)
	if v == nil {
		return nil, ErrNotExist
	}
	return v, nil
}

const delimiter = "||"

// Format formats the context with only builtin keys
func Format(ctx context.Context) (result string, finalErr error) {
	defer func() {
		if e := recover(); nil != e {
			finalErr = bytes.ErrTooLarge
		}
	}()
	buf := new(bytes.Buffer)
	for _, key := range builtInKeys {
		val, err := getVal(ctx, key)
		if nil != err {
			continue
		}
		kname, ok := keyName[key]
		if !ok {
			continue
		}
		if shouldConvert2Duration(key) {
			i64 := val.(*int64)
			val = time.Duration(*i64)
		}
		_, finalErr = buf.WriteString(fmt.Sprintf(kname+"=%v"+delimiter, val))
	}
	return strings.TrimRight(buf.String(), delimiter), nil
}

// TraceString is the same as Format,but only return string
// that is, if there's an error occurs, TraceString will return "ctx_format=unset"
func TraceString(ctx context.Context) string {
	if ctx == nil {
		return "ctx=null"
	}
	s, err := Format(ctx)
	if nil != err || "" == s {
		return "ctx_format=unset"
	}
	return s
}
