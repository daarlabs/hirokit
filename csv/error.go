package csv

import "errors"

var (
	ErrorTargetNotPtr = errors.New("target must not be a pointer")
	ErrorTargetPtr    = errors.New("target must be a pointer")
	ErrorTargetSlice  = errors.New("target must be a slice")
)
