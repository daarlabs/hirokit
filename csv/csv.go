package csv

import (
	"reflect"
	"strings"
	"time"
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
