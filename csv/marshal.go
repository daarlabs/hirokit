package csv

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	
	"github.com/daarlabs/hirokit/util/convertx"
)

func (m *manager) Marshal(data any) ([]byte, error) {
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	result := make([]string, 0)
	if t.Kind() == reflect.Ptr {
		return []byte{}, ErrorTargetNotPtr
	}
	if t.Kind() != reflect.Slice {
		return []byte{}, ErrorTargetSlice
	}
	result = append(result, strings.Join(m.fields, string(m.separator)))
	rt := t.Elem()
	for i := 0; i < v.Len(); i++ {
		row := v.Index(i)
		if row.Kind() != reflect.Struct {
			continue
		}
		resultRow := make([]string, len(m.fields))
		for j := 0; j < row.NumField(); j++ {
			fv := row.Field(j)
			ft := rt.Field(j)
			tag := ft.Tag.Get(csvTag)
			if tag == "-" || !fv.IsValid() {
				continue
			}
			var fieldName string
			var fieldIndex int
			for fi, fn := range m.fields {
				if tag == fn {
					fieldName = tag
					fieldIndex = fi
					break
				}
			}
			if len(fieldName) == 0 {
				continue
			}
			switch ft.Type.Kind() {
			case reflect.Struct:
				switch ft.Type {
				case timeType:
					resultRow[fieldIndex] = convertx.ConvertTimeToString(fv.Interface().(time.Time))
					continue
				}
				resultRow[fieldIndex] = convertx.ConvertSqlNullToString(fv.Interface())
			default:
				resultRow[fieldIndex] = fmt.Sprintf("%v", fv.Interface())
			}
		}
		result = append(result, strings.Join(resultRow, string(m.separator)))
	}
	return []byte(strings.Join(result, "\r\n") + "\r\n"), nil
}

func (m *manager) MustMarshal(data any) []byte {
	b, err := m.Marshal(data)
	if err != nil {
		panic(err)
	}
	return b
}
