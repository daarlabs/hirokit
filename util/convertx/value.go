package convertx

import (
	"reflect"
	"strconv"
	"strings"
	
	"github.com/daarlabs/hirokit/errs"
)

var (
	floatReplacer = strings.NewReplacer(
		",", ".",
		" ", "",
	)
)

func ConvertValue(src string, t interface{}) error {
	tv := reflect.ValueOf(t)
	if tv.Kind() != reflect.Ptr {
		return errs.ErrorPointerTarget
	}
	switch reflect.TypeOf(t).Elem().Kind() {
	case reflect.Int:
		val, err := strconv.Atoi(src)
		if err != nil {
			return err
		}
		tv.Elem().Set(reflect.ValueOf(val))
	case reflect.Int64:
		val, err := strconv.ParseInt(src, 10, 64)
		if err != nil {
			return err
		}
		tv.Elem().Set(reflect.ValueOf(val))
	case reflect.Float32:
		val, err := strconv.ParseFloat(floatReplacer.Replace(src), 32)
		if err != nil {
			return err
		}
		tv.Elem().Set(reflect.ValueOf(val))
	case reflect.Float64:
		val, err := strconv.ParseFloat(floatReplacer.Replace(src), 64)
		if err != nil {
			return err
		}
		tv.Elem().Set(reflect.ValueOf(val))
	case reflect.Bool:
		val, err := strconv.ParseBool(src)
		if err != nil {
			return err
		}
		tv.Elem().Set(reflect.ValueOf(val))
	case reflect.String:
		tv.Elem().Set(reflect.ValueOf(src))
	default:
		return errs.ErrorUnsupportedType
	}
	return nil
}

func MustConvertValue(src string, t interface{}) {
	if err := ConvertValue(src, t); err != nil {
		panic(err)
	}
}
