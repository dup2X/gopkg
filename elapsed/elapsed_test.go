package elapsed

import (
	"testing"
	"time"
)

func TestElapsed(t *testing.T) {
	timeNow = func() time.Time {
		return time.Now().Add(1 * time.Second)
	}

	et := New()
	et.Start()
	time.Sleep(time.Second * 1)
	elapsed := et.Stop()
	if elapsed < time.Duration(time.Second) {
		t.FailNow()
	}

	et.Reset()
	et.Start()
	time.Sleep(time.Second * 1)
	timeNow = time.Now
	elapsed = et.Stop()
	if elapsed >= time.Duration(time.Second) {
		t.FailNow()
	}
	println(et.Elapsed().String())
}
