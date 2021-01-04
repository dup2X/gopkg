package idgen

import (
	"testing"
)

func TestNextID(t *testing.T) {
	ig := New("123.12.2.34")
	id, err := ig.NextID()
	if err != nil {
		t.FailNow()
	}
	id1, err := ig.NextID()
	if err != nil {
		t.FailNow()
	}
	if id1 <= id {
		t.FailNow()
	}
}
