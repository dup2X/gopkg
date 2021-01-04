package idl

import (
	"fmt"
	"strconv"
)

func toint(v interface{}) int {
	i, _ := strconv.Atoi(tostr(v))
	return i
}

func tostr(v interface{}) string {
	return fmt.Sprint(v)
}
