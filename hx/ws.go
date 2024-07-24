package hx

import "github.com/daarlabs/hirokit/gox"

func Ws(url string) gox.Node {
	return gox.Fragment(
		Ext("ws"),
		gox.CreateAttribute[string]("ws-connect")(url),
	)
}
