// Package logger ...
package logger

import (
	"testing"

	"github.com/dup2X/gopkg/config"
)

func TestOfflineFileLog(t *testing.T) {
	cfg, err := config.New("./testdata/test.conf")
	if err != nil {
		t.FailNow()
	}

	sec, err := cfg.GetSection("offline")
	f := newOfflineFileLog(sec)
	f.track("public", "a=1||b=c")
	f.track("public", "a=1||b=2")
	f.track("public", "a=1||b=3")
	f.track("public", "a=1||b=4")
}
