package context

import (
	stdctx "context"
	"fmt"
	"testing"
	"time"

	dctx "github.com/dup2X/gopkg/ctxutil"
)

func TestTraceInfo(t *testing.T) {
	ctx := stdctx.TODO()
	ctx = SetAPPTraceInfo(ctx, "start")
	AppendAPPTraceInfo(ctx, "step1")
	AppendAPPTraceInfo(ctx, "step2")
	AppendAPPTraceInfo(ctx, "stop")
	rets := GetAPPTraceInfo(ctx)
	fmt.Printf("%+v\n", rets)
}

func TestCheckSLA(t *testing.T) {
	ctx := stdctx.TODO()
	ctx = dctx.SetRequestInTs(ctx, time.Now().UnixNano()/1e6)
	ctx = dctx.SetRequestTimeout(ctx, 1009)
	time.Sleep(time.Second)
	fmt.Println(CheckSLA(ctx))
}
