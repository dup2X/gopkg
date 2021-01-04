package logger

import (
	"testing"
)

func TestFileLogWithOption(t *testing.T) {
	log := NewLoggerWithOption("/tmp/log", SetPrefix("test_option"), SetRotateByHour(), SetAutoClearHours(2))
	log.Trace(123)
	log.Tracef("%s--%d", "test", nowFunc().Unix())
	log.Debug(123)
	log.Debugf("%s--%d", "test", nowFunc().Unix())
	log.Info(123)
	log.Infof("%s--%d", "test", nowFunc().Unix())
	log.Warn(123)
	log.Warnf("%s--%d", "test", nowFunc().Unix())
	log.Error(123)
	log.Errorf("%s--%d", "test", nowFunc().Unix())
}
