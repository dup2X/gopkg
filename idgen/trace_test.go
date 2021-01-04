package idgen

import (
	"fmt"
	"testing"
	"time"
)

func TestGenTraceID(t *testing.T) {
	err := NewTraceGen()
	if err != nil {
		t.FailNow()
	}
	traceID := GenTraceID()
	println(traceID)
	traceID = GenTraceID()
	println(traceID)
}

func BenchmarkLoops(b *testing.B) {
	err := NewTraceGen()
	if err != nil {
		b.FailNow()
	}
	for i := 0; i < b.N; i++ {
		GenTraceID()
	}
}
func BenchmarkLoopsParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		err := NewTraceGen()
		if err != nil {
			b.FailNow()
		}
		for pb.Next() {
			GenTraceID()
		}
	})
}

func Test_ParseUnixNanoTimeToSecond(t *testing.T) {
	timestamp := time.Now().UnixNano()
	seconds := parseUnixNanoTimeToSecond(timestamp)
	fmt.Printf("seconds = %d\n", seconds)
}

func Test_ParseUnixNanoTimeToMilSecond(t *testing.T) {
	timestamp := time.Now().UnixNano()
	milseconds := parseUnixNanoTimeToMilSecond(timestamp)
	fmt.Printf("seconds = %d\n", milseconds)
}

func Test_GenSpanID(t *testing.T) {
	spanID := GenSpanID()
	fmt.Printf("spanId = %s, len = %d \n", spanID, len(spanID))
	if len(spanID) != 16 {
		t.Errorf("len = %d, spanId = %s", len(spanID), spanID)
	}
}
