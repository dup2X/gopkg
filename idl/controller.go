// Package idl ...
package idl

import (
	"context"
)

// IController ...
type IController interface {
	GetRequestIDL() interface{}
	Do(context.Context, interface{}) (interface{}, APIErr)
}
