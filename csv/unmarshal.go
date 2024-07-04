package csv

import (
	"reflect"
	
	"github.com/daarlabs/hirokit/util/convertx"
)

func (m *manager) Unmarshal(b []byte, target any) error {
	_, err := m.Read(b)
	if err != nil {
		return err
	}
	v := reflect.ValueOf(target)
	t := reflect.TypeOf(target)
	if t.Kind() != reflect.Ptr {
		return ErrorTargetPtr
	}
	if t.Elem().Kind() != reflect.Slice {
		return ErrorTargetSlice
	}
	sliceType := t.Elem()
	targetType := t.Elem().Elem()
	targetSlice := reflect.MakeSlice(sliceType, 0, len(m.rows))
	fieldsLen := len(m.fields)
	for _, row := range m.rows {
		targetRow := reflect.New(targetType)
		for i, col := range row {
			if i >= fieldsLen {
				continue
			}
			fieldName := m.fields[i]
			for j := 0; j < targetType.NumField(); j++ {
				ft := targetType.Field(j)
				fv := targetRow.Elem().Field(j)
				tag := ft.Tag.Get(csvTag)
				if tag == "" || tag == "-" || !fv.IsValid() || !fv.CanAddr() || !fv.CanSet() {
					continue
				}
				if tag != fieldName {
					continue
				}
				switch ft.Type.Kind() {
				case reflect.Struct:
					switch ft.Type {
					case timeType:
						fv.Set(reflect.ValueOf(convertx.ConvertStringToTime(col)))
						continue
					}
					fv.Set(reflect.ValueOf(convertx.ConvertStringToSqlNull(col, ft.Type)))
				default:
					if err := convertx.ConvertValue(col, fv.Addr().Interface()); err != nil {
						return err
					}
				}
			}
		}
		targetSlice = reflect.Append(targetSlice, targetRow.Elem())
	}
	v.Elem().Set(targetSlice)
	return nil
}

func (m *manager) MustUnmarshal(b []byte, target any) {
	if err := m.Unmarshal(b, target); err != nil {
		panic(err)
	}
}
