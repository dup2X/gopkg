package errors

import (
	"testing"
)

func TestAll(t *testing.T) {
	err := New("level 1")
	err = Wrap(err, "level 2")
	if Cause(err).Error() != "level 1" {
		t.Fatal(err.Error())
	}
	if err.Error() != "level 2: level 1" {
		t.Fatal(err.Error())
	}
	println(PrintStack(err))
	err = Errorf("base %t", true)
	err = Wrapf(err, "format:%d", 1)
	println(err.Error())
}
