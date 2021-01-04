// Package httpsvr ...
package httpsvr

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dup2X/gopkg/idl"
	"github.com/dup2X/gopkg/logger"
)

var defaultTimeout = 2000
var defaultMarshalFn = func(v interface{}, err idl.APIErr) ([]byte, error) {
	ret := map[string]interface{}{
		"errno":  err.Code(),
		"errmsg": err.Error(),
		"data":   v,
	}
	return json.Marshal(ret)
}

type option struct {
	dumpResponse  bool
	enableElasped bool
	dumpAccess    bool
	validate      bool
	log           logger.Logger
	marshalFunc   func(ret interface{}, err idl.APIErr) ([]byte, error)
	unmarshalFunc func(*http.Request, interface{}) error
	readTimeout   time.Duration
	writeTimeout  time.Duration
	handleTimeout int64
}

//ServerOption 定义ServerOption类型
type ServerOption func(o *option)

//WithDumpResponse dump response
func WithDumpResponse(dump bool) ServerOption {
	return func(o *option) {
		o.dumpResponse = dump
	}
}

//EnableElasped enable elapsed
func EnableElasped(enable bool) ServerOption {
	return func(o *option) {
		o.enableElasped = enable
	}
}

//WithDumpAccess dump access
func WithDumpAccess(dump bool) ServerOption {
	return func(o *option) {
		o.dumpAccess = dump
	}
}

//WithLogger log
func WithLogger(log logger.Logger) ServerOption {
	return func(o *option) {
		o.log = log
	}
}

//EnableValidate enable validate
func EnableValidate(enable bool) ServerOption {
	return func(o *option) {
		o.validate = enable
	}
}

// SetLocalLanguage ...
func SetLocalLanguage(lang string) ServerOption {
	return func(o *option) {
		localLanguage = lang
	}
}

// SetServerReadTimeout ...
func SetServerReadTimeout(rt time.Duration) ServerOption {
	return func(o *option) {
		o.readTimeout = rt
	}
}

// SetServerWriteTimeout ...
func SetServerWriteTimeout(wt time.Duration) ServerOption {
	return func(o *option) {
		o.writeTimeout = wt
	}
}

// SetHandleDefaultTimeout ...
func SetHandleDefaultTimeout(to int64) ServerOption {
	return func(o *option) {
		o.handleTimeout = to
	}
}

//SetServerMarshalFunc ...
func SetServerMarshalFunc(marshalFunc func(v interface{}, err idl.APIErr) ([]byte, error)) ServerOption {
	return func(o *option) {
		o.marshalFunc = marshalFunc
	}
}

// SetServerUnmarshalFunc ...
func SetServerUnmarshalFunc(fn func(*http.Request, interface{}) error) ServerOption {
	return func(o *option) {
		o.unmarshalFunc = fn
	}
}
