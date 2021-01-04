package main

import (
	"time"

	"github.com/dup2X/gopkg/httpsvr"
	"github.com/dup2X/gopkg/httpsvr/_example/ctrls"
)

func main() {
	s := httpsvr.New("127.0.0.1:10024",
		httpsvr.SetServerReadTimeout(time.Millisecond*200),
		httpsvr.SetServerWriteTimeout(time.Millisecond*200),
		httpsvr.SetHandleDefaultTimeout(2000),
	)
	s.AddRoute("POST", "/test/api", &ctrls.DemoControler{}, httpsvr.SetHandleTimeout(3000))
	s.Serve()
}
