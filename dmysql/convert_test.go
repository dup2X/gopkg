package dmysql

import (
	"testing"
)

type Test struct {
	Name   string `json:"name"`
	Age    int    `json:"age"`
	Level  int64  `json:"level,omitempty"`
	NoTag  string
	ignore int
}

func TestDecodeRowMaps(t *testing.T) {
	var as = []interface{}{
		&Test{},
		&Test{},
	}

	var ms = []RowMap{
		map[string]string{
			"name":   "test",
			"age":    "11",
			"level":  "",
			"NoTag":  "no",
			"ignore": "2",
		},
		map[string]string{
			"name":   "test1",
			"age":    "111",
			"level":  "11",
			"NoTag":  "no1",
			"ignore": "21",
		},
	}
	err := DecodeRowMaps(as, ms)
	if err != nil {
		t.FailNow()
	}
	a, ok := as[0].(*Test)
	if !ok {
		t.FailNow()
	}
	if a.Age != 11 || a.Name != "test" || a.Level != 0 {
		t.FailNow()
	}
	a1, ok := as[1].(*Test)
	if !ok {
		t.FailNow()
	}
	if a1.Age != 111 || a1.Name != "test1" || a1.Level != 11 {
		t.FailNow()
	}
}
