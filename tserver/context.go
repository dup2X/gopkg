package tserver

import (
	stdctx "context"
	"time"

	"github.com/dup2X/gopkg/context"
)

type thriftContext struct {
	stdctx.Context
	context.Header
}

func (tc *thriftContext) GetTraceID() string { return tc.Header.TraceID }

func (tc *thriftContext) GetSpanID() string { return tc.Header.SpanID }

func (tc *thriftContext) GetHintCode() int { return tc.Header.HintCode }

func (tc *thriftContext) GetHintContent() string { return tc.Header.HintContent }

func (tc *thriftContext) GetElapseds() map[string]time.Duration { return tc.Header.Elapsed }

func (tc *thriftContext) AddElapsed(key string, elapsed time.Duration) {
	if tc.Header.Elapsed == nil {
		tc.Header.Elapsed = make(map[string]time.Duration)
	}
	tc.Header.Elapsed[key] += elapsed
}
