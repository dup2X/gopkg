// Package context ...
package context

import (
	stdctx "context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"

	dctx "github.com/dup2X/gopkg/ctxutil"
)

var (
	traceKey        = "#gopkg.trace#"
	degradeKey      = "degrade_settings"
	redisElapsedKey = "redis_elapsed"
	mysqlElapsedKey = "mysql_elapsed"
	rpcElapsedKey   = "rpc_elapsed"
	callerKey       = "caller"
)

// SetTrace set trace_info with http request
func SetTrace(ctx stdctx.Context, r *http.Request) {
	hd := &Header{
		TraceID:     r.Header.Get(TraceIDKey),
		SpanID:      r.Header.Get(SpanIDKey),
		HintContent: r.Header.Get(HintContentKey),
		Elapsed:     make(map[string]time.Duration),
	}
	if hd.TraceID == "" {
		hd.TraceID = r.Header.Get(phpTraceIDKey)
	}
	if hd.SpanID == "" {
		hd.SpanID = r.Header.Get(phpSpanIDKey)
	}
	hd.HintCode, _ = strconv.Atoi(r.Header.Get(HintCodeKey))
	ctx = stdctx.WithValue(ctx, traceKey, hd)
}

// GetTrace ...
func GetTrace(ctx stdctx.Context) *Header {
	v := ctx.Value(traceKey)
	if v == nil {
		return genHeader()
	}
	hd, ok := v.(*Header)
	if !ok {
		return genHeader()
	}
	return hd
}

// TraceString dump trace info
func TraceString(ctx stdctx.Context) string {
	if ctx == nil {
		return "ctx=nil"
	}
	v := ctx.Value(traceKey)
	if v == nil {
		return "trace_id=unset"
	}
	hd, ok := v.(*Header)
	if !ok {
		return "trace_id=unexpected"
	}
	return hd.TraceID
}

// CheckSLA return whether time is enough for feature proc
// if HintContent set it
func CheckSLA(ctx stdctx.Context) bool {
	if ctx == nil {
		return true
	}
	tStart, err := dctx.GetRequestInTs(ctx)
	if err != nil {
		return true
	}
	tOut, err := dctx.GetRequestTimeout(ctx)
	if err != nil {
		return true
	}
	tNow := time.Now().UnixNano() / 1e6
	return (tNow - tStart) < tOut
}

// former version
/*func CheckSLA(ctx stdctx.Context) bool {
	if ctx == nil {
		return true
	}
	v := ctx.Value(traceKey)
	if v == nil {
		return true
	}
	hd, ok := v.(*Header)
	if !ok {
		return true
	}
	return hd.CheckSLA()
}*/

// SetRedisElapsed ...
func SetRedisElapsed(ctx stdctx.Context) stdctx.Context {
	var redisElapsed = int64(0)
	return stdctx.WithValue(ctx, redisElapsedKey, &redisElapsed)
}

// AddRedisElapsed add some elapsed to ctx
func AddRedisElapsed(ctx stdctx.Context, cost time.Duration) {
	if ctx == nil {
		return
	}
	v := ctx.Value(redisElapsedKey)
	if v == nil {
		return
	}
	hd, ok := v.(*int64)
	if !ok {
		return
	}
	atomic.AddInt64(hd, int64(cost))
}

// GetRedisElapsed ...
func GetRedisElapsed(ctx stdctx.Context) string {
	v := ctx.Value(redisElapsedKey)
	if v == nil {
		return "0"
	}
	hd, ok := v.(*int64)
	if !ok {
		return "0"
	}
	return time.Duration(*hd).String()
}

// SetMysqlElapsed ...
func SetMysqlElapsed(ctx stdctx.Context) stdctx.Context {
	var mysqlElapsed = int64(0)
	return stdctx.WithValue(ctx, mysqlElapsedKey, &mysqlElapsed)
}

// AddMysqlElapsed add some elapsed to ctx
func AddMysqlElapsed(ctx stdctx.Context, cost time.Duration) {
	if ctx == nil {
		return
	}
	v := ctx.Value(mysqlElapsedKey)
	if v == nil {
		return
	}
	hd, ok := v.(*int64)
	if !ok {
		return
	}
	atomic.AddInt64(hd, int64(cost))
}

// GetMysqlElapsed ...
func GetMysqlElapsed(ctx stdctx.Context) string {
	v := ctx.Value(mysqlElapsedKey)
	if v == nil {
		return "0"
	}
	hd, ok := v.(*int64)
	if !ok {
		return "0"
	}
	return time.Duration(*hd).String()
}

// SetDegrade ...
func SetDegrade(ctx stdctx.Context, sed int) stdctx.Context {
	return stdctx.WithValue(ctx, degradeKey, sed)
}

// GetDegrade ...
func GetDegrade(ctx stdctx.Context) int {
	v := ctx.Value(degradeKey)
	if v == nil {
		return 0
	}
	if val, ok := v.(int); ok {
		return val
	}
	return 0
}

// SetRPCElapsed ...
func SetRPCElapsed(ctx stdctx.Context, rpcStr fmt.Stringer) stdctx.Context {
	return stdctx.WithValue(ctx, rpcElapsedKey, rpcStr)
}

// GetRPCElapsed ...
func GetRPCElapsed(ctx stdctx.Context) fmt.Stringer {
	v := ctx.Value(rpcElapsedKey)
	if v == nil {
		return nilRPCString
	}
	if val, ok := v.(fmt.Stringer); ok {
		return val
	}
	return nilRPCString
}

// SetCaller ...
func SetCaller(ctx stdctx.Context, uri *url.URL) stdctx.Context {
	return stdctx.WithValue(ctx, callerKey, uri.Path)
}

// GetCaller ...
func GetCaller(ctx stdctx.Context) string {
	v := ctx.Value(callerKey)
	if v == nil {
		return ""
	}
	if val, ok := v.(string); ok {
		return val
	}
	return ""
}

var nilRPCString = &nilStr{}

type nilStr struct{}

func (n *nilStr) String() string {
	return "rpc=0"
}
