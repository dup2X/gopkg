// Package httpsvr ...
package httpsvr

import (
	"net/http"

	"github.com/dup2X/gopkg/idl"
)

type ctrlOption struct {
	handleTimeout     int64
	unmarshalFunc     func(r *http.Request, req interface{}) error
	marshalFunc       func(v interface{}, err idl.APIErr) ([]byte, error)
	addResponseHeader func() http.Header
}

// ControllerOption 定义ControllerOption类型
type ControllerOption func(o *ctrlOption)

// SetControllerMarshalFunc ...
func SetControllerMarshalFunc(fn func(v interface{}, err idl.APIErr) ([]byte, error)) ControllerOption {
	return func(o *ctrlOption) {
		o.marshalFunc = fn
	}
}

//SetControllerUnmarshalFunc 设置某个Controller的Unmarshal方法
func SetControllerUnmarshalFunc(fn func(r *http.Request, req interface{}) error) ControllerOption {
	return func(o *ctrlOption) {
		o.unmarshalFunc = fn
	}
}

// AddResponseHeader ...
func AddResponseHeader(fn func() http.Header) ControllerOption {
	return func(o *ctrlOption) {
		o.addResponseHeader = fn
	}
}

// SetHandleTimeout ...
func SetHandleTimeout(to int64) ControllerOption {
	return func(o *ctrlOption) {
		o.handleTimeout = to
	}
}
