package hx

import "github.com/daarlabs/hirokit/gox"

const (
	SwapInnerHtml   = "innerHTML"
	SwapOuterHtml   = "outerHTML"
	SwapBeforeBegin = "beforebegin"
	SwapAfterBegin  = "afterbegin"
	SwapBeforeEnd   = "beforeend"
	SwapAfterEnd    = "afterend"
	SwapDelete      = "delete"
	SwapNone        = "none"
)

func Swap(value ...string) gox.Node {
	return gox.CreateAttribute[string](atrributePrefix + "-swap")(value...)
}

func Append() gox.Node {
	return gox.CreateAttribute[string](atrributePrefix + "-swap")(SwapBeforeEnd)
}
