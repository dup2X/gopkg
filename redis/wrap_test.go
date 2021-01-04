package redis

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	dctx "github.com/dup2X/gopkg/context"
)

func TestWrap(t *testing.T) {
	var (
		strVal     = "aaa"
		strsVal    = []string{"aaa", "bbbb"}
		intVal     = 123
		intsVal    = []int{1, 2, 3}
		int64Val   = int64(1 << 33)
		bytesVal   = []byte("hello")
		byteSlices = [][]byte{[]byte("hello"), []byte("world")}
		ctx        = context.Background()
	)
	r, _ := http.NewRequest("GET", "aa", nil)
	dctx.SetTrace(ctx, r)
	c, err := genConn()
	assert(t, err == nil)
	c.Set(ctx, "str", strVal)
	c.Set(ctx, "strs", strsVal)
	c.Set(ctx, "int", intVal)
	c.Set(ctx, "ints", intsVal)
	c.Set(ctx, "int64", int64Val)
	c.Set(ctx, "bytes", bytesVal)
	c.Set(ctx, "byteSlices", byteSlices)

	v1, err := String(c.Get(ctx, "str"))
	assert(t, err == nil && v1 == strVal)
	v2, err := Int(c.Get(ctx, "int"))
	assert(t, err == nil && v2 == intVal)
	v3, err := Int64(c.Get(ctx, "int64"))
	assert(t, err == nil && v3 == int64Val)
	Strings(c.Get(ctx, "strs"))

	v5, err := Bytes(c.Get(ctx, "bytes"))
	assert(t, err == nil && reflect.DeepEqual(v5, bytesVal))
}
