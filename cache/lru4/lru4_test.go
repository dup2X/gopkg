package lru4

import (
	"fmt"
	"testing"
)

func TestLRU4(t *testing.T) {
	c := New(10)
	for i := 0; i < 14; i++ {
		c.Add(fmt.Sprintf("key-%d", i), i)
	}
	for i := 0; i < 14; i++ {
		_, ok := c.Get(fmt.Sprintf("key-%d", i))
		if ok && i < 4 {
			t.FailNow()
		}
		if !ok && i > 3 {
			t.FailNow()
		}
	}
}
