package devtool

import . "github.com/daarlabs/hirokit/gox"

type Plugin struct {
	Title    string
	IconPath Node
	RowFunc  func(value string) Node
}

const (
	PluginDebug    = "debug"
	PluginDatabase = "database"
	PluginSession  = "session"
	PluginCache    = "cache"
)
