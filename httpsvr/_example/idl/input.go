package idl

// RequestSt ...
type RequestSt struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

// ResponseSt ...
type ResponseSt struct {
	Count int64 `json:"count"`
	Hit   int64 `json:"hit"`
}
