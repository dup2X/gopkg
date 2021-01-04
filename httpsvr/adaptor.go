// Package httpsvr ...
package httpsvr

import (
	stdctx "context"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/dup2X/gopkg/context"
	"github.com/dup2X/gopkg/ctxutil"
	"github.com/dup2X/gopkg/idgen"
	"github.com/dup2X/gopkg/idl"
)

var (
	errBindFailed    = errors.New("input data bind idl failed, please check your request")
	errNoMarshalFunc = errors.New("no marshal func")
)

type httpAdapt struct {
	handleTimeout     int64
	unmarshalFunc     func(r *http.Request, req interface{}) error
	marshalFunc       func(v interface{}, err idl.APIErr) ([]byte, error)
	addResponseHeader func() http.Header
}

//Accept 接受请求
func (ha *httpAdapt) Accept(r *http.Request, req interface{}) (stdctx.Context, error) {
	ctx := stdctx.TODO()
	ctx = setTrace(ctx, r)
	ctx = idgen.SetLogID(ctx, idgen.GenLogID(r))
	ctx = ctxutil.SetDegrade(ctx, rand.Intn(100))
	ctx = ctxutil.SetCaller(ctx, r.URL.Path)
	ctx = context.SetRedisElapsed(ctx)
	ctx = context.SetMysqlElapsed(ctx)
	ctx = ctxutil.SetHTTPRequest(ctx, r)
	ctx = ctxutil.SetRequestInTs(ctx, time.Now().UnixNano()/1e6)
	ctx = ctxutil.SetRequestTimeout(ctx, ha.getTimeout(r))
	ctx = ctxutil.SetCookies(ctx, r.Cookies())
	return ctx, ha.unmarshalFunc(r, req)
}

func (ha *httpAdapt) getTimeout(r *http.Request) int64 {
	tw := r.Header.Get(context.RPCTimeoutMsKey)
	if tw == "" {
		return ha.handleTimeout
	}
	i, err := strconv.ParseInt(tw, 10, 64)
	if err != nil || i < 1 {
		return ha.handleTimeout
	}
	return i
}

func setTrace(ctx stdctx.Context, r *http.Request) stdctx.Context {
	ctx = context.SetAPPTraceInfo(ctx, "set_trace")
	traceID := r.Header.Get(context.TraceIDKey)
	if traceID == "" {
		traceID = idgen.GenTraceID()
	}
	spanID := r.Header.Get(context.SpanIDKey)
	if spanID == "" {
		spanID = idgen.GenSpanID()
	}
	hcode := r.Header.Get(context.HintCodeKey)
	hintCode, _ := strconv.ParseInt(hcode, 10, 64)
	hcontent := r.Header.Get(context.HintContentKey)
	ctx = ctxutil.SetTraceID(ctx, traceID)
	ctx = ctxutil.SetSpanID(ctx, spanID)
	ctx = ctxutil.SetHintCode(ctx, hintCode)
	ctx = ctxutil.SetHintContent(ctx, hcontent)
	return ctx
}

func newHTTPAdapter(options *ctrlOption, sopt *option) *httpAdapt {
	adp := &httpAdapt{}
	adp.setOptions(options)
	if adp.unmarshalFunc == nil {
		adp.unmarshalFunc = sopt.unmarshalFunc
	}
	if adp.marshalFunc == nil {
		adp.marshalFunc = sopt.marshalFunc
	}
	if adp.handleTimeout == 0 {
		adp.handleTimeout = sopt.handleTimeout
	}
	return adp
}

func (ha *httpAdapt) setOptions(options *ctrlOption) {
	if options == nil {
		return
	}
	if options.unmarshalFunc != nil {
		ha.unmarshalFunc = options.unmarshalFunc
	}
	if options.marshalFunc != nil {
		ha.marshalFunc = options.marshalFunc
	}
	if options.addResponseHeader != nil {
		ha.addResponseHeader = options.addResponseHeader
	}
	if options.handleTimeout != 0 {
		ha.handleTimeout = options.handleTimeout
	}
}
