package httpsvr

type okErr struct {
	code int
	error
}

func (n *okErr) Code() int {
	return 0
}

func (n *okErr) Error() string {
	return "ok"
}
