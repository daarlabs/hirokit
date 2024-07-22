package escapex

import "strings"

var (
	urlEscaper = strings.NewReplacer(
		`"`, ``,
		`'`, ``,
	)
)

func Url(v string) string {
	return urlEscaper.Replace(v)
}
