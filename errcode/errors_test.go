package errcode

import "testing"

func GetTestVar() map[BizErr]string {
	t := map[BizErr]string{
		ErrCommonParamInvalidType:       "param type is error",
		ErrCommonParamInvalidValue:      "param value is error",
		ErrCommonParamJSONEncodeGetNull: "json encode param get a null",
		ErrCommonParamJSONEncodeFail:    "json encode data fail",
	}
	return t
}

func TestPassengerErr_Code(t *testing.T) {
	tArr := GetTestVar()
	for tKey := range tArr {
		ret := BizErr(tKey.Code())
		if ret != tKey {
			t.Errorf("PassengerErr Code failed. Got %d, expected %d", ret, tKey)
		}
	}
}

func TestPassengerErr_Error(t *testing.T) {
	tArr := GetTestVar()
	for tKey, tVal := range tArr {
		ret := tKey.Error()
		if ret != tVal {
			t.Errorf("PassengerErr Error failed. Got %s, expected %s", ret, tVal)
		}
	}
}
