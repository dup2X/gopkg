// package main ...
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/dup2X/gopkg/config"
	"github.com/dup2X/gopkg/logger"
)

func test() {
	log := logger.NewLoggerWithOption("/tmp/log",
		logger.SetPrefix("test_option"), logger.SetRotateByHour(), logger.SetAutoClearHours(2))
	log.Trace("trace")
	log.Debug("trace")
	log.Info("trace")
	log.Warn("trace")
	log.Error("trace")
	log.Debugf("trace=%d", 2)
}

func main() {
	if false {
		test()
		return
	}
	cfg, err := config.New("../testdata/test.conf")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	err = logger.NewLoggerWithConfig(cfg)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	ctx := context.TODO()
	for i := 0; i < 10; i++ {
		logger.Trace(logger.DLTagUndefined, "trace")
		logger.Debug(logger.DLTagUndefined, "trace")
		logger.Info(logger.DLTagUndefined, "trace")
		logger.Warn(logger.DLTagUndefined, "trace")
		logger.Error(logger.DLTagUndefined, "trace")
		logger.Debugf(ctx, logger.DLTagUndefined, "trace=%d", i)
	}
	time.Sleep(time.Second * 2)

	// 打离线日志
	trackSec, err := cfg.GetSection("offline")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	logger.InitTrack(trackSec)
	logger.Track("public", "xxxxxxxx") // 往public.log 写入
	logger.Track("track", "xxxxxxxx")  // 往 trace.log写入
}
