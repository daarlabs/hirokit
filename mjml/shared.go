package mjml

import "github.com/daarlabs/hirokit/gox"

func Title(nodes ...gox.Node) gox.Node {
	return gox.CreateShared("title")(nodes...)
}
