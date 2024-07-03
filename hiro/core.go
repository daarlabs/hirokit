package hiro

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	
	"github.com/daarlabs/hirokit/config"
	"github.com/daarlabs/hirokit/logger"
)

type Mirage interface {
	Router
	Log(handler logger.Handler) Mirage
	DynamicHandler(handler Handler) Mirage
	Layout() LayoutManager
	Run(address string)
	Mux() *http.ServeMux
	Plugin(plugin Plugin) Mirage
}

type core struct {
	*router
	*assets
	config         config.Config
	dynamicHandler Handler
	layout         *layout
	mux            *http.ServeMux
	plugins        []Plugin
	routes         []*Route
}

const (
	logo = `
    __  __________  ____
   / / / /  _/ __ \/ __ \
  / /_/ // // /_/ / / / /
 / __  // // _, _/ /_/ /
/_/ /_/___/_/ |_|\____/
`
)

const (
	Version = "0.1.0"
)

func New(cfg config.Config) Mirage {
	cfg = cfg.Init()
	mux := http.NewServeMux()
	rts := make([]*Route, 0)
	c := &core{
		config:         cfg,
		dynamicHandler: defaultDynamicHandler,
		layout:         createLayout(),
		mux:            mux,
		routes:         rts,
		plugins:        make([]Plugin, 0),
	}
	c.assets = createAssets(cfg)
	c.router = &router{
		config: cfg,
		mux:    mux,
		prefix: cfg.Router.Prefix,
		routes: &rts,
		assets: c.assets,
	}
	c.router.core = c
	c.router.createDynamicAssetsRoute()
	c.router.createWildcardRoute()
	c.onInit()
	return c
}

func (c *core) Log(handler logger.Handler) Mirage {
	if c.config.Logger == nil {
		return c
	}
	c.config.Logger.HandleFunc(handler)
	return c
}

func (c *core) DynamicHandler(handler Handler) Mirage {
	c.dynamicHandler = handler
	return c
}

func (c *core) Layout() LayoutManager {
	return c.layout
}

func (c *core) Run(address string) {
	if strings.HasPrefix(address, ":") {
		address = "localhost" + address
	}
	fmt.Println(logo)
	fmt.Println("")
	fmt.Println("Name: ", c.config.App.Name)
	fmt.Println("Address: ", address)
	fmt.Println("Version: ", Version)
	for _, p := range c.plugins {
		if len(p.Name) == 0 {
			continue
		}
		fmt.Println("Plugin loaded: ", p.Name)
	}
	log.Fatalln(http.ListenAndServe(address, c.mux))
}

func (c *core) Mux() *http.ServeMux {
	return c.mux
}

func (c *core) Plugin(plugin Plugin) Mirage {
	c.plugins = append(c.plugins, plugin)
	if c.config.Localization.Translator != nil {
		for langCode, locales := range plugin.Locales {
			c.config.Localization.Translator.Extend(langCode, locales)
		}
	}
	return c
}

func (c *core) onInit() {
	c.assets.mustProcess()
}