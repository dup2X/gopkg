// Package httpsvr ...
package httpsvr

import (
	"errors"
	"io/ioutil"
	"net/http"
	"sync/atomic"

	"github.com/julienschmidt/httprouter"
)

const (
	traceIDKey = "HTTP_HEADER_RID"
	spanIDKey  = "HTTP_HEADER_SPANID"

	phpTraceIDKey  = "header-rid"
	phpSpanIDKey   = "header-spanid"
	hintCodeKey    = "hintCode"
	hintContentKey = "hintContent"
	languageKey    = "lang"
)

var (
	// ErrEmptyBody body为空的错误消息
	ErrEmptyBody = errors.New("empty request body")
)

// Context 上下文结构，TraceID、SpanID、ResponseWriter、Request、Params等
type Context struct {
	Params       httprouter.Params
	Req          *http.Request
	TraceID      string
	SpanID       string
	ParentSpanID string

	written uint32
	w       http.ResponseWriter
}

func newContext(w http.ResponseWriter, r *http.Request, params httprouter.Params) *Context {
	return &Context{
		Params:  params,
		Req:     r,
		TraceID: r.Header.Get(traceIDKey),
		SpanID:  r.Header.Get(spanIDKey),
		w:       w,
	}
}

func (c *Context) refill() {
	if c.TraceID == "" {
		c.TraceID = c.Req.Header.Get(phpTraceIDKey)
	}
	if c.SpanID == "" {
		c.SpanID = c.Req.Header.Get(phpSpanIDKey)
	}
}

// Write 写响应数据
func (c *Context) Write(data []byte, code int) {
	if atomic.LoadUint32(&c.written) == 1 {
		// TODO
		return
	}
	c.w.WriteHeader(code)
	c.w.Write(data)
}

//FormValues 解析参数并格式化为map
func (c *Context) FormValues() map[string]string {
	c.Req.ParseForm()
	ret := make(map[string]string)
	for k, vs := range c.Req.Form {
		if vs != nil && len(vs) > 0 {
			ret[k] = vs[0]
		}
	}
	return ret
}

//GetRequestBody 获取请求体
func (c *Context) GetRequestBody() ([]byte, error) {
	if c.Req.Body == nil {
		return nil, ErrEmptyBody
	}
	data, err := ioutil.ReadAll(c.Req.Body)
	if err != nil {
		return nil, err
	}
	return data, c.Req.Body.Close()
}

//TraceHeader header中的traceID、spanID和parentSpanID
type TraceHeader struct {
	traceID      string
	spanID       string
	parentSpanID string
}

//NewTraceHeader 创建TraceHeader
func NewTraceHeader(r *http.Request) *TraceHeader {
	return &TraceHeader{
		traceID: r.Header.Get(traceIDKey),
		spanID:  r.Header.Get(spanIDKey),
	}
}

//GetTraceID 从TraceHeader获取traceID
func (th *TraceHeader) GetTraceID() string {
	return th.traceID
}

//GetSpanID 从TraceHeader获取spanID
func (th *TraceHeader) GetSpanID() string {
	return th.spanID
}

//GetParentSpanID 从TraceHeader获取parentSpanID
func (th *TraceHeader) GetParentSpanID() string {
	return th.parentSpanID
}

//Encode 将traceID和spanID格式化为map
func (th *TraceHeader) Encode() map[string]string {
	return map[string]string{
		traceIDKey: th.traceID,
		spanIDKey:  th.spanID,
	}
}
