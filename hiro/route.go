package hiro

import (
	"regexp"
	
	"github.com/daarlabs/hirokit/firewall"
	"github.com/daarlabs/hirokit/socketer"
)

type RouteConfig struct {
	Type  int
	Value any
}

type Route struct {
	Lang       string
	Path       string
	Name       string
	Layout     layoutFactory
	Matcher    *regexp.Regexp
	Methods    []string
	Firewall   []firewall.Firewall
	PathValues []string
	Ws         socketer.Ws
	Components map[string]MandatoryComponent
}

const (
	routeMethod = iota
	routeName
	routeLayout
	routeWs
)

func Method(method ...string) RouteConfig {
	return RouteConfig{
		Type:  routeMethod,
		Value: method,
	}
}

func Name(name string) RouteConfig {
	return RouteConfig{
		Type:  routeName,
		Value: name,
	}
}

func Layout(name string) RouteConfig {
	return RouteConfig{
		Type:  routeLayout,
		Value: name,
	}
}
