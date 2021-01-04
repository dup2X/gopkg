package idl

// Request ...
type Request interface {
	GetValidateDef() ValidateDef
}

// Response ...
type Response interface {
	Encode() ([]byte, error)
}

// APIErr ...
type APIErr interface {
	error
	Code() int
}
