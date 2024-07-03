package hx

import "github.com/daarlabs/hirokit/gox"

func Trigger(value ...string) gox.Node {
	return gox.CreateAttribute[string](atrributePrefix + "-trigger")(value...)
}
