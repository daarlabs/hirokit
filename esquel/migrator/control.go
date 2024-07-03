package migrator

import "github.com/daarlabs/hirokit/esquel"

type Control interface {
	DB(name ...string) *esquel.DB
}

type control struct {
	*migrator
}

func (c *control) DB(name ...string) *esquel.DB {
	n := mainDbname
	if len(name) > 0 {
		n = name[0]
	}
	d, ok := c.databases[n]
	if !ok {
		panic(ErrorInvalidDatabase)
	}
	return d
}
