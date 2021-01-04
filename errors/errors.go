package errors

// See https://github.com/pkg/errors for more information
// How to handle err - http://dave.cheney.net/paste/gocon-spring-2016.pdf
// File errors_test.go shows how to use

import (
	"fmt"

	"github.com/pkg/errors"
)

//New : new error
func New(msg string) error {
	return errors.New(msg)
}

//Errorf : 通过参数格式化error
func Errorf(format string, args ...interface{}) error {
	return errors.Errorf(format, args...)
}

//Wrap : 包装error
func Wrap(err error, msg string) error {
	return errors.Wrap(err, msg)
}

//Wrapf : 包装并且格式化error
func Wrapf(err error, format string, args ...interface{}) error {
	return errors.Wrapf(err, format, args...)
}

//Cause : cause error
func Cause(err error) error {
	return errors.Cause(err)
}

//PrintStack : 打印error栈
func PrintStack(err error) string {
	return fmt.Sprintf("%+v", err)
}
