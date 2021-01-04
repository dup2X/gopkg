// Package httpsvr ...
package httpsvr

import (
	"fmt"
)

//Response 响应结构体
type Response struct {
	Code int         `json:"errno"`
	Msg  string      `json:"errmsg"`
	Data interface{} `json:"data"`
}

//GenNotImplementResponse 产生404 响应
func GenNotImplementResponse(name string) *Response {
	return &Response{
		Code: 404,
		Msg:  fmt.Sprintf("%s has not be implemented!", name),
	}
}

//GenResponse 根据参数产生响应
func GenResponse(data interface{}, code int, msg string) *Response {
	return &Response{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

//GenExtraResponse 根据map产生响应
func GenExtraResponse(extData map[string]interface{}, code int, msg string) interface{} {
	if extData == nil {
		extData = make(map[string]interface{})
	}
	extData["errno"] = code
	extData["msg"] = msg
	return extData
}

var shortSLAResponse = []byte(`{"errno":499,"errmsg":"insufficient time-balance","data":{}}`)
