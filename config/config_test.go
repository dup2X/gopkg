package config

import (
	"runtime"
	"testing"
)

var testcases = []struct {
	sec      string
	strKey   string
	strVal   string
	intKey   string
	intVal   int64
	boolKey  string
	boolVal  bool
	floatKey string
	floatVal float64
}{
	{
		sec:      "test",
		strKey:   "str",
		strVal:   "strVal",
		intKey:   "int",
		intVal:   1024,
		boolKey:  "bool",
		boolVal:  true,
		floatKey: "float",
		floatVal: 2.45,
	},
}

func assert(t *testing.T, exp bool) {
	if !exp {
		_, f, n, _ := runtime.Caller(1)
		println("err@", f, n)
		t.FailNow()
	}
}
