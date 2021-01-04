package metrics

import (
	"net"
	"net/url"
)

// AddHTTPResponseErr 处理http response err
func AddHTTPResponseErr(err error, moduleName, funcName string) {
	keyPrefix := moduleName + "_"
	if funcName != "" {
		keyPrefix += funcName + "_"
	}
	keyPrefix += "request_"
	switch myerr := err.(type) {
	case net.Error:
		if myerr.Timeout() {
			Add(keyPrefix+"timeout", 1)
		} else {
			Add(keyPrefix+"neterr", 1)
		}
	case *url.Error:
		if myerr.Timeout() {
			Add(keyPrefix+"timeout", 1)
		} else {
			Add(keyPrefix+"urlerr", 1)
		}
	default:
		Add(keyPrefix+"failed", 1)
	}
}
