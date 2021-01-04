package errcode

//BizErr :Biz module err type
type BizErr int

/**
* 通用错误码 520000 ~ 529999
 */
// [520000 ~ 520200) 000  Param
const (
	//ErrCommonParamInvalidType:520000
	ErrCommonParamInvalidType BizErr = 520000 + iota
	//ErrCommonParamInvalidValue:520001
	ErrCommonParamInvalidValue
	//ErrCommonParamJSONEncodeGetNull:520002
	ErrCommonParamJSONEncodeGetNull
	//ErrCommonParamJSONEncodeFail:520003
	ErrCommonParamJSONEncodeFail
	//ErrCommonParamJSONDecodeFail:520004
	ErrCommonParamJSONDecodeFail
	//ErrCommonParamCheckSignFail:520005
	ErrCommonParamCheckSignFail
)

//Code :get error code
func (pErr BizErr) Code() int {
	return (int)(pErr)
}

//Error :get error msg
func (pErr BizErr) Error() string {
	switch pErr {
	case ErrCommonParamInvalidType:
		return "param type is error"
	case ErrCommonParamInvalidValue:
		return "param value is error"
	case ErrCommonParamJSONEncodeGetNull:
		return "json encode param get a null"
	case ErrCommonParamJSONEncodeFail:
		return "json encode data fail"
	case ErrCommonParamJSONDecodeFail:
		return "json decode data fail"
	case ErrCommonParamCheckSignFail:
		return "check sign data fail"

	default:
		return "unknow err"
	}
}
