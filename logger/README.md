# logger
日志模块

## feture
* 多输出 － 同时支持标准输出，文件，socket等
* 配置驱动
* 文件切割 － 基于时间和文件大小

## example
test.conf
```
[log]
type = stdout,file // 打标准输出和文件
prefix = xxx // 日志文件前缀 xxx.log.xxxx

file.enable = true
file.dir = /xxx/log
file.rotate_by_hour = true
"file.rotate_size = 10240000
file.level = TRACE
file.format = [%L][%Z][%S]%M // 日志输出格式
file.seprated = true // 如果为true则将info以上日志写进xxx.log.wf，其他写进xxx.log; false=所有级别在一个文件
file.auto_clear = true
file.clear_hours = 72

stdout.enable = false
"stdout.format = [%t %d] [%L] %M
"stdout.format = [%t %d] %M
stdout.level = TRACE

[offline]
dir = /tmp/log
file_list = public,track // 写两个文件，一个public.log 一个track.log
rotate_by_hour=true

```

demo.go
```
import (
	"github.com/dup2X/gopkg/config"
	"github.com/dup2X/gopkg/logger"
)

func main() {
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
		logger.Debugf(ctx, "trace=%d", logger.DLTagUndefined, i)
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
	...
}

```
