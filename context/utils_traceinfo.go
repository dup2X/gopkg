// Package context ...
package context

import (
	"container/list"
	stdctx "context"
	"fmt"
)

var appTraceInfoKey = "APP_TRACE_INFO"

// SetAPPTraceInfo ...
func SetAPPTraceInfo(ctx stdctx.Context, info string) stdctx.Context {
	linkList := list.New()
	linkList.PushBack(info)
	return stdctx.WithValue(ctx, appTraceInfoKey, linkList)
}

// AppendAPPTraceInfo ...
func AppendAPPTraceInfo(ctx stdctx.Context, info string) {
	v := ctx.Value(appTraceInfoKey)
	if v != nil {
		if link, ok := v.(*list.List); ok {
			link.PushBack(info)
		}
	}
}

// GetAPPTraceInfo ...
func GetAPPTraceInfo(ctx stdctx.Context) []string {
	v := ctx.Value(appTraceInfoKey)
	if v == nil {
		return nil
	}
	link, ok := v.(*list.List)
	if !ok {
		return nil
	}
	h := link.Front()
	var rets []string
	for {
		if h == nil {
			break
		}
		rets = append(rets, fmt.Sprint(h.Value))
		h = h.Next()
	}
	return rets
}
