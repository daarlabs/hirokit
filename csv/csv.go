package csv

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	
	"github.com/daarlabs/hirokit/util/convertx"
)

type Manager interface {
	Fields(fields ...string) Manager
	Separator(s rune) Manager
	Read(b []byte) (int, error)
	Marshal(data any) ([]byte, error)
	Unmarshal(b []byte, target any) error
	
	MustRead(b []byte) int
	MustMarshal(data any) []byte
	MustUnmarshal(b []byte, target any)
}

type manager struct {
	b         []byte
	fields    []string
	rows      [][]string
	separator rune
}

const (
	csvTag = "csv"
)

const (
	valueMethod = "Value"
)

var (
	timeType = reflect.TypeOf(time.Time{})
)

func New() Manager {
	return &manager{
		separator: ';',
	}
}

func (m *manager) Fields(fields ...string) Manager {
	m.fields = fields
	return m
}

func (m *manager) Separator(s rune) Manager {
	m.separator = s
	return m
}

func (m *manager) Read(b []byte) (int, error) {
	m.b = b[:]
	var fieldsCount int
	m.fields = make([]string, 0)
	m.rows = make([][]string, 0)
	parsedRows := strings.Split(string(m.b), "\r\n")
	for i, parsedRow := range parsedRows {
		cols := strings.Split(strings.TrimSpace(parsedRow), string(m.separator))
		n := len(cols)
		if i == 0 {
			fieldsCount = n
			for _, c := range cols {
				m.fields = append(m.fields, strings.TrimSpace(c))
			}
			continue
		}
		if fieldsCount != n {
			continue
		}
		m.rows = append(m.rows, cols)
	}
	return len(b), nil
}

func (m *manager) MustRead(b []byte) int {
	n, err := m.Read(b)
	if err != nil {
		panic(err)
	}
	return n
}

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
		resultRow := make([]string, row.NumField())
		for _, fieldName := range m.fields {
			for j := 0; j < row.NumField(); j++ {
				fv := row.Field(j)
				ft := rt.Field(j)
				tag := ft.Tag.Get(csvTag)
				if tag != fieldName {
					continue
				}
				switch ft.Type.Kind() {
				case reflect.Struct:
					switch ft.Type {
					case timeType:
						resultRow[j] = fmt.Sprintf("%s", fv.Interface().(time.Time).UTC().String())
						continue
					}
					vm := fv.MethodByName(valueMethod)
					if !vm.IsValid() || (vm.IsValid() && vm.Type().NumIn() > 0) {
						continue
					}
					callResult := vm.Call(nil)
					if len(callResult) == 0 {
						continue
					}
					resultRow[j] = fmt.Sprintf("%v", callResult[0])
				default:
					resultRow[j] = fmt.Sprintf("%v", fv.Interface())
				}
			}
		}
		result = append(result, strings.Join(resultRow, string(m.separator)))
	}
	return []byte(strings.Join(result, "\r\n") + "\r\n"), nil
}

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
				if tag == "" || !fv.IsValid() || !fv.CanAddr() || !fv.CanSet() {
					continue
				}
				if tag != fieldName {
					continue
				}
				if err := convertx.ConvertValue(col, fv.Addr().Interface()); err != nil {
					return err
				}
			}
		}
		targetSlice = reflect.Append(targetSlice, targetRow.Elem())
	}
	v.Elem().Set(targetSlice)
	return nil
}

func (m *manager) MustMarshal(data any) []byte {
	b, err := m.Marshal(data)
	if err != nil {
		panic(err)
	}
	return b
}

func (m *manager) MustUnmarshal(b []byte, target any) {
	if err := m.Unmarshal(b, target); err != nil {
		panic(err)
	}
}
