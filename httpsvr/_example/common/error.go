package common

import (
	"github.com/dup2X/gopkg/errcode"
	"github.com/dup2X/gopkg/idl"
)

var _ idl.APIErr = new(DemoErr)

// DemoErr ...
type DemoErr struct {
	code errcode.RespCode
	err  error
}

// Error ...
func (dr *DemoErr) Error() string {
	if dr.err != nil {
		return dr.err.Error()
	}
	return "ok"
}

// Code ...
func (dr *DemoErr) Code() int {
	return int(dr.code)
}

// GenErr ...
func GenErr(code errcode.RespCode, err error) *DemoErr {
	return &DemoErr{code: code, err: err}
}
