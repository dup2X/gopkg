// Package logger ...

// +build go1.7

package logger

import (
	"context"
	"testing"

	"github.com/dup2X/gopkg/config"
)

func TestFileLog(t *testing.T) {
	ctx := context.TODO()
	cfg, err := config.New("./testdata/test.conf")
	if err != nil {
		t.FailNow()
	}
	NewLoggerWithConfig(cfg)
	Trace(DLTagUndefined, 123)
	Tracef(ctx, DLTagUndefined, "%s--%d", "test", nowFunc().Unix())
	Debug(DLTagUndefined, 123)
	Debugf(ctx, DLTagUndefined, "%s--%d", "test", nowFunc().Unix())
	Info(DLTagUndefined, 123)
	Infof(ctx, DLTagUndefined, "%s--%d", "test", nowFunc().Unix())
	Warn(DLTagUndefined, 123)
	Warnf(ctx, DLTagUndefined, "%s--%d", "test", nowFunc().Unix())
	Error(DLTagUndefined, 123)
	Errorf(ctx, DLTagUndefined, "%s--%d", "test", nowFunc().Unix())
	Public(ctx, "xxx", map[string]interface{}{"a": 1, "v": true})
	stack := PrintStack()
	if stack == nil || len(stack) < 1 {
		t.FailNow()
	}
}
