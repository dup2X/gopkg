// Package httpsvr ...
package httpsvr

import (
	"strconv"
)

// APIResponseHeader api响应头格式
type APIResponseHeader interface {
	GetCode() int
	GetMsg() string
}

// PHPResponseHeader php响应头格式
type PHPResponseHeader struct {
	Code interface{} `json:"errno"`
	Msg  string      `json:"errmsg"`
}

// GetCode 获取PHPResponseHeader响应码
func (ph *PHPResponseHeader) GetCode() int {
	switch ph.Code.(type) {
	case int:
		return ph.Code.(int)
	case int64:
		return int(ph.Code.(int64))
	case string:
		i, _ := strconv.Atoi(ph.Code.(string))
		return i
	default:
		return -1
	}
}

//GetMsg 获取PHPResponseHeader响应消息
func (ph *PHPResponseHeader) GetMsg() string {
	return ph.Msg
}
