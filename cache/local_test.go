package cache

import (
	"testing"
)

var testcase = []struct {
	key, val []byte
}{
	{
		key: []byte("k1"),
		val: []byte("v1"),
	},
	{
		key: []byte("k2"),
		val: []byte("v2"),
	},
	{
		key: []byte("k3"),
		val: []byte("v3"),
	},
}

func TestLocalSet(t *testing.T) {
	loc := NewLocal()
	for i := range testcase {
		loc.Set(testcase[i].key, testcase[i].val)
	}
}
func TestLocalGet(t *testing.T) {
	loc := NewLocal()
	for i := range testcase {
		loc.Set(testcase[i].key, testcase[i].val)
	}
	for i := range testcase {
		if string(testcase[i].val) != string(loc.Get(testcase[i].key)) {
			t.FailNow()
		}
	}
}

func TestLocalDel(t *testing.T) {
	loc := NewLocal()
	for i := range testcase {
		loc.Set(testcase[i].key, testcase[i].val)
	}
	for i := range testcase {
		if string(testcase[i].val) != string(loc.Get(testcase[i].key)) {
			t.FailNow()
		}
	}
	loc.Del(testcase[0].key)
	if loc.Get(testcase[0].key) != nil {
		t.FailNow()
	}
}

func TestLocalFlushAll(t *testing.T) {
	loc := NewLocal()
	for i := range testcase {
		loc.Set(testcase[i].key, testcase[i].val)
	}
	for i := range testcase {
		if string(testcase[i].val) != string(loc.Get(testcase[i].key)) {
			t.FailNow()
		}
	}
	loc.FlushAll()
	for i := range testcase {
		if nil != loc.Get(testcase[i].key) {
			t.FailNow()
		}
	}
}

func BenchmarkSet(b *testing.B) {
	loc := NewLocal()
	for i := 0; i < b.N; i++ {
		loc.Set(testcase[0].key, testcase[0].val)
	}
}

func BenchmarkGet(b *testing.B) {
	loc := NewLocal()
	loc.Set(testcase[0].key, testcase[0].val)
	for i := 0; i < b.N; i++ {
		loc.Get(testcase[0].key)
	}

}

func BenchmarkDel(b *testing.B) {
	loc := NewLocal()
	loc.Set(testcase[0].key, testcase[0].val)
	for i := 0; i < b.N; i++ {
		loc.Del(testcase[0].key)
	}

}

func BenchmarkFlushAll(b *testing.B) {
	loc := NewLocal()
	loc.Set(testcase[0].key, testcase[0].val)
	for i := 0; i < b.N; i++ {
		loc.FlushAll()
	}
}
