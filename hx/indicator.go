package hx

import "github.com/daarlabs/hirokit/gox"

func Indicator(value ...string) gox.Node {
	return gox.CreateAttribute[string](atrributePrefix + "-indicator")(value...)
}
