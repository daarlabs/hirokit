package esquel

import "errors"

var (
	ErrorMismatchArgs = errors.New("placeholders and args count mismatch")
)
