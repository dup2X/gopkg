// Package dmysql ...
package dmysql

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	defaultTag = "json"
)

var (
	errNilRowMap      = fmt.Errorf("nil RowMap cannot be decode")
	errNotMatchedData = fmt.Errorf("RowMaps has different length with struct slice")
)

// DecodeRowMap use reflect decode RowMap into struct
func DecodeRowMap(target interface{}, r RowMap) error {
	if r == nil {
		return errNilRowMap
	}
	v := reflect.ValueOf(target)
	if v.Type().Kind() != reflect.Ptr || v.Elem().Type().Kind() != reflect.Struct {
		panic("DecodeRow should decode struct-pointer")
	}
	t := reflect.TypeOf(target).Elem()
	l := v.Elem().NumField()
	for i := 0; i < l; i++ {
		f := t.Field(i)
		vf := v.Elem().Field(i)
		if !vf.CanSet() {
			continue
		}
		var (
			str, name string
			lst       []string
		)
		str = f.Tag.Get(defaultTag)
		if str == "" {
			name = f.Name
		} else {
			lst = strings.Split(str, ",")
			if len(lst) < 1 {
				continue
			}
			name = lst[0]
		}
		val, ok := r[name]
		if !ok {
			continue
		}
		if f.Type.Kind() == reflect.String {
			vf.SetString(val)
			continue
		}
		switch f.Type.Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32:
			val, err := strconv.ParseInt(r[name], 10, 64)
			if err != nil {
				val = 0
			}
			vf.SetInt(val)
		case reflect.Float64, reflect.Float32:
			val, err := strconv.ParseFloat(r[name], 64)
			if err != nil {
				val = 0.00
			}
			vf.SetFloat(val)
		}
	}
	return nil
}

// DecodeRowMaps ...
func DecodeRowMaps(sts []interface{}, rs []RowMap) error {
	if rs == nil || sts == nil {
		return errNilRowMap
	}
	if len(rs) != len(sts) {
		return errNotMatchedData
	}
	for i := range rs {
		if err := DecodeRowMap(sts[i], rs[i]); err != nil {
			return err
		}
	}
	return nil
}
