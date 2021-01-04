package idl

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dup2X/gopkg/logger"

	json "github.com/bitly/go-simplejson"
)

var errBind = errors.New("bind err")

// JSON ...
const JSON = "json"

// Bind ...
func Bind(r *http.Request, req Request) error {
	// TODO
	return BindAndValidate(r, req)
}

// BindAndValidate ...
func BindAndValidate(r *http.Request, req Request) error {
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "form-urlencoded") {
		return form(r, req)
	}

	//	if strings.Contains(contentType, "json") {
	//		return json(r, req)
	//	}
	return nil

}

// form ...
func form(r *http.Request, req Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	vd := req.GetValidateDef()
	for k, v := range vd {
		val, ok := r.Form[v.JSON]
		if !ok {
			if !validateAndBind(v, k, nil) {
				return fmt.Errorf("not exist:%s", v.JSON)
			}
			continue
		}
		logger.Debugf(nil, logger.DLTagUndefined, "bind and validate faild:%v %v %v", v, k, val[0])
		if !validateAndBind(v, k, val[0]) {
			return fmt.Errorf("bind and validate faild:%v %v %v", v, k, val[0])
		}
	}
	return nil
}

// validateAndBind ...
func validateAndBind(f *Field, k, v interface{}) bool {
	if f.Required && v == nil {
		return false
	}
	if !f.Required && v == nil {
		v = f.Default
	}
	switch t := k.(type) {
	case *string:
		var val string
		switch o := v.(type) {
		case *json.Json:
			val, _ = o.String()
		default:
			val = tostr(v)
		}
		if f.MaxLen > 0 && len(val) > f.MaxLen {
			logger.Errorf(nil, logger.DLTagUndefined, "len(%v) is out range [%v,%v]", val, f.MinLen, f.MaxLen)
			return false
		}
		*t = val
	case *int:
		var val int
		switch o := v.(type) {
		case *json.Json:
			val, _ = o.Int()
		default:
			val = toint(v)
		}
		if val < toint(f.MinVal) || val > toint(f.MaxVal) {
			logger.Errorf(nil, logger.DLTagUndefined, "%v is out range [%v,%v]", val, f.MinVal, f.MaxVal)
			return false
		}
		*t = val
		logger.Debugf(nil, logger.DLTagUndefined, "%s intval=%d", f.JSON, val)
	default:
		xx, ok := k.(IStruct)
		if ok {
			switch f.Codec {
			case JSON:
				if err := validateAndBindForJSON([]byte(tostr(v)), xx); err != nil {
					logger.Errorf(nil, logger.DLTagUndefined, "codec failed,err:%v src:%s", err, tostr(v))
					return false
				}
			}
		} else {
			println("not IStruct")
		}
		t = xx
	}
	return true
}

func validateAndBindForJSON(data []byte, is IStruct) error {
	obj, err := json.NewJson(data)
	if err != nil {
		return err
	}
	for k1, v1 := range is.GetValidateDef() {
		v, ok := obj.CheckGet(v1.JSON)
		if !ok {
			if !validateAndBind(v1, k1, nil) {
				return errBind
			}
			continue
		}
		if !validateAndBind(v1, k1, v) {
			return errBind
		}
	}
	return nil
}
