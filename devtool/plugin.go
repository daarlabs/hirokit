package devtool

import . "github.com/daarlabs/hirokit/gox"

type Plugin struct {
	Title     string
	IconPath  Node
	Reference bool
	RowFunc   func(value string) Node
}

const (
	PluginDebug    = "debug"
	PluginDatabase = "database"
	PluginSession  = "session"
	PluginParam    = "param"
)
