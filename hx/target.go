package hx

import "github.com/daarlabs/hirokit/gox"

func Target(value ...string) gox.Node {
	return gox.CreateAttribute[string](atrributePrefix + "-target")(value...)
}
