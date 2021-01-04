// Package idl ...
package idl

// ValidateDef ...
type ValidateDef map[interface{}]*Field

// Field ...
type Field struct {
	Required bool
	JSON     string
	Default  interface{}
	MinLen   int
	MaxLen   int
	MaxVal   interface{}
	MinVal   interface{}
	Codec    string
}

// IStruct ...
type IStruct interface {
	GetValidateDef() ValidateDef
}
