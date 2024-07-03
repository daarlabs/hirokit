package form

import "bytes"

type Multipart struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Suffix string `json:"suffix"`
	Data   []byte `json:"data"`
}

func (m Multipart) Read(p []byte) (n int, err error) {
	return bytes.NewBuffer(m.Data).Read(p)
}

func (m Multipart) Write(p []byte) (n int, err error) {
	return bytes.NewBuffer(m.Data).Write(p)
}
