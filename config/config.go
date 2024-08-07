package config

import (
	"github.com/daarlabs/hirokit/esquel"
	"github.com/daarlabs/hirokit/filesystem"
	"github.com/daarlabs/hirokit/form"
	"github.com/daarlabs/hirokit/logger"
	"github.com/daarlabs/hirokit/mailer"
	"github.com/daarlabs/hirokit/socketer"
)

type Config struct {
	App          App
	Cache        Cache
	Database     map[string]*esquel.DB
	Dev          Dev
	Export       Export
	Form         form.Config
	Filesystem   filesystem.Config
	Localization Localization
	Logger       *logger.Logger
	Parser       Parser
	Router       Router
	Security     Security
	Smtp         mailer.Config
	Ws           socketer.Config
}

func (c Config) Init() Config {
	if c.Form.Limit == 0 {
		c.Form.Limit = form.DefaultBodyLimit
	}
	return c
}
