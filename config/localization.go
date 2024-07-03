package config

import (
	"github.com/daarlabs/hirokit/translator"
	
	"github.com/daarlabs/hirokit/form"
)

type Localization struct {
	Enabled    bool
	Path       bool
	Languages  []Language
	Translator translator.Translator
	Form       form.Messages
}

type Language struct {
	Main bool
	Code string
}
