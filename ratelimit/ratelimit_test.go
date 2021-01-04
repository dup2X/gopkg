package ratelimit

import (
	"testing"
	"time"
)

//TestGetTicket :test
func TestGetTicket(t *testing.T) {
	l := New(40000)
	ticket := l.GetTicket()
	if ticket != true {
		t.FailNow()
	}
	lf := NewWithFillInterval(time.Millisecond*20, 10000)
	lf.GetTicket()
}
