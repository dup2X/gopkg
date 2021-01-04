package ctrls

import (
	"context"
	"fmt"

	"github.com/dup2X/gopkg/ctxutil"
	"github.com/dup2X/gopkg/httpsvr/_example/handlers"
	"github.com/dup2X/gopkg/httpsvr/_example/idl"
	sidl "github.com/dup2X/gopkg/idl"
)

// DemoControler ...
type DemoControler struct {
}

// GetRequestIDL ...
func (dc *DemoControler) GetRequestIDL() interface{} {
	return &idl.RequestSt{}
}

// Do ...
func (dc *DemoControler) Do(ctx context.Context, req interface{}) (interface{}, sidl.APIErr) {
	sla, _ := ctxutil.GetRequestTimeout(ctx)
	fmt.Println("!!!!!!!", sla)
	rreq := req.(*idl.RequestSt)
	// TODO 做一些基本处理
	return handlers.DemoProc(rreq)
}
