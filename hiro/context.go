package hiro

import (
	"context"
	"net/http"
	"sync"
	
	"github.com/daarlabs/hirokit/esquel"
	"github.com/daarlabs/hirokit/filesystem"
	"github.com/daarlabs/hirokit/logger"
	
	"github.com/daarlabs/hirokit/auth"
	"github.com/daarlabs/hirokit/cache"
	"github.com/daarlabs/hirokit/config"
	"github.com/daarlabs/hirokit/cookie"
	"github.com/daarlabs/hirokit/csrf"
	"github.com/daarlabs/hirokit/mailer"
	"github.com/daarlabs/hirokit/parser"
	"github.com/daarlabs/hirokit/sender"
)

type Ctx interface {
	Auth(dbname ...string) auth.Manager
	Cache() cache.Client
	Config() config.Config
	Continue() error
	Cookie() cookie.Cookie
	Create() Factory
	Csrf() csrf.Csrf
	DB(dbname ...string) *esquel.DB
	Email() mailer.Mailer
	Export() Export
	Files() filesystem.Client
	Flash() Flash
	Generate() Generator
	Lang() Lang
	Log() logger.Log
	Page() Page
	Parse() parser.Parse
	Request() Request
	Response() Response
	State() State
	Translate(key string, args ...map[string]any) string
}

type ctx struct {
	context.Context
	err              error
	cachedComponents *map[string]MandatoryComponent
	config           config.Config
	cookie           cookie.Cookie
	csrf             csrf.Csrf
	files            filesystem.Client
	mu               *sync.Mutex
	page             *page
	r                *http.Request
	w                http.ResponseWriter
	route            *Route
	routes           *[]*Route
	response         *response
	state            *state
	assets           *assets
	lang             *lang
	component        *componentCtx
	write            *bool
	parsed           Map
}

type ctxParam struct {
	assets           *assets
	cachedComponents *map[string]MandatoryComponent
	config           config.Config
	layout           *layout
	matchedRoute     *Route
	routes           *[]*Route
	r                *http.Request
	w                http.ResponseWriter
}

func createContext(p ctxParam) *ctx {
	cx := context.Background()
	write := true
	c := &ctx{
		Context:          cx,
		cachedComponents: p.cachedComponents,
		config:           p.config,
		files:            filesystem.New(cx, p.config.Filesystem),
		mu:               &sync.Mutex{},
		page:             createPage(),
		route:            p.matchedRoute,
		routes:           p.routes,
		r:                p.r,
		w:                p.w,
		assets:           p.assets,
		write:            &write,
		parsed:           make(Map),
	}
	c.cookie = cookie.New(c.r, c.w, c.createCookiePathBasedOnRouterCookiePrefix())
	if c.config.Security.Csrf != nil && c.config.Security.Csrf.IsEnabled() {
		c.csrf = csrf.New(
			csrf.Cache(c.Cache()),
			csrf.Cookie(c.cookie),
			csrf.Request(p.r),
			csrf.Expiration(c.config.Security.Csrf.GetExpiration()),
		)
	}
	if c.config.Localization.Enabled {
		c.lang = createLang(c.Config(), c.Request(), c.Cookie())
	}
	c.response = &response{
		Sender: sender.New(p.r, p.w, &write),
		ctx:    c,
		route:  p.matchedRoute,
	}
	c.state = createState(c.Cache(), c.Cookie())
	c.parsePathValues()
	return c
}

func (c *ctx) Auth(dbname ...string) auth.Manager {
	var db *esquel.DB
	var ok bool
	dbn := Main
	if len(dbname) > 0 {
		dbn = dbname[0]
	}
	if len(c.config.Database) > 0 {
		db, ok = c.config.Database[dbn]
		if !ok {
			panic(ErrorInvalidDatabase)
		}
	}
	return auth.New(
		db,
		c.r,
		c.w,
		c.cookie,
		c.Cache(),
		c.config.Security.Auth,
	)
}

func (c *ctx) Cache() cache.Client {
	return cache.New(c.Context, c.config.Cache.Memory, c.config.Cache.Redis)
}

func (c *ctx) Config() config.Config {
	return c.config
}

func (c *ctx) Cookie() cookie.Cookie {
	return c.cookie
}

func (c *ctx) Continue() error {
	// c.mu.Unlock()
	return nil
}

func (c *ctx) Create() Factory {
	return factory{ctx: c}
}

func (c *ctx) Csrf() csrf.Csrf {
	return c.csrf
}

func (c *ctx) DB(dbname ...string) *esquel.DB {
	dbn := Main
	if len(dbname) > 0 {
		dbn = dbname[0]
	}
	db, ok := c.config.Database[dbn]
	if !ok {
		panic(ErrorInvalidDatabase)
	}
	return db
}

func (c *ctx) Email() mailer.Mailer {
	return mailer.New(c.config.Smtp)
}

func (c *ctx) Export() Export {
	return createExport(c.config.Export)
}

func (c *ctx) Files() filesystem.Client {
	return c.files
}

func (c *ctx) Flash() Flash {
	return flash{state: c.state}
}

func (c *ctx) Generate() Generator {
	return &generator{c}
}

func (c *ctx) Lang() Lang {
	return c.lang
}

func (c *ctx) Log() logger.Log {
	return c.config.Logger
}

func (c *ctx) Page() Page {
	return c.page
}

func (c *ctx) Parse() parser.Parse {
	return parser.New(c.r, c.config.Parser.Limit)
}

func (c *ctx) Request() Request {
	return request{
		r:            c.r,
		route:        c.route,
		componentCtx: c.component,
		parsed:       c.parsed,
	}
}

func (c *ctx) Response() Response {
	return c.response
}

func (c *ctx) State() State {
	return c.state
}

func (c *ctx) Translate(key string, args ...map[string]any) string {
	if !c.config.Localization.Enabled {
		return key
	}
	return c.config.Localization.Translator.Translate(c.Lang().Current(), key, args...)
}

func (c *ctx) createCookiePathBasedOnRouterCookiePrefix() string {
	if len(c.config.Router.Prefix.Cookie) > 0 {
		return c.config.Router.Prefix.Cookie
	}
	return "/"
}

func (c *ctx) parsePathValues() {
	if c.route == nil {
		return
	}
	for _, pathValueKey := range c.route.PathValues {
		c.parsed[pathValueKey] = c.r.PathValue(pathValueKey)
	}
}
