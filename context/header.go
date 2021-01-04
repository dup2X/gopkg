// Package context ...
package context

import (
	"time"

	"github.com/bitly/go-simplejson"
)

// Header ...
type Header struct {
	// TraceID ...
	TraceID string
	// SpanID ...
	SpanID string
	// HintCode ...
	HintCode int
	// HintContent ...
	HintContent string
	// Elapsed ...
	Elapsed map[string]time.Duration
}

// CheckSLA ...
func (h Header) CheckSLA() bool {
	obj, err := simplejson.NewJson([]byte(h.HintContent))
	if err != nil {
		return true
	}
	bl, ok := obj.CheckGet("balance")
	if !ok {
		return true
	}
	if bl.MustInt() == 0 {
		return false
	}
	return true
}

func genHeader() *Header {
	// TODO
	return &Header{}
}
