package config

import (
	"github.com/daarlabs/hirokit/auth"
	"github.com/daarlabs/hirokit/csrf"
	"github.com/daarlabs/hirokit/firewall"
)

type Security struct {
	Auth     auth.Config
	Csrf     csrf.Csrf
	Firewall []firewall.Firewall
}
