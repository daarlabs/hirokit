package hiro

import (
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strings"
	
	"github.com/daarlabs/hirokit/config"
	"github.com/daarlabs/hirokit/constant/fileSuffix"
	"github.com/daarlabs/hirokit/constant/header"
	"github.com/daarlabs/hirokit/firewall"
	"github.com/daarlabs/hirokit/tempest"
	"github.com/daarlabs/hirokit/util/pathx"
)

type Router interface {
	Static(path, dir string) Router
	Route(path any, handler Handler, config ...RouteConfig) Router
	Group(path any, name ...string) Router
}

type router struct {
	core        *core
	config      config.Config
	mux         *http.ServeMux
	prefix      config.Prefix
	middlewares []Handler
	assets      *assets
	routes      *[]*Route
}

const (
	paramRegex = `[0-9a-zA-Z]+`
)

const (
	dynamicName = "dynamic"
)

func (r *router) Static(path, dir string) Router {
	r.mux.Handle(http.MethodGet+" "+path, http.StripPrefix(path, http.FileServer(http.Dir(dir))))
	return r
}

func (r *router) Route(path any, fn Handler, config ...RouteConfig) Router {
	switch v := path.(type) {
	case string:
		if r.config.Localization.Path {
			for _, item := range r.config.Localization.Languages {
				r.createRoute(v, fn, item.Code, config...)
				r.createCanonicalRoute(v, item.Code, config...)
			}
		}
		if !r.config.Localization.Path {
			r.createRoute(v, fn, "", config...)
			r.createCanonicalRoute(v, "", config...)
		}
	case map[string]string:
		for l, p := range v {
			r.createRoute(p, fn, l, config...)
			r.createCanonicalRoute(p, "", config...)
		}
	}
	return r
}

func (r *router) Group(path any, name ...string) Router {
	var routerName string
	if len(name) > 0 {
		routerName = name[0]
	}
	return &router{
		core:   r.core,
		config: r.config,
		mux:    r.mux,
		prefix: config.Prefix{
			Path: r.mergePrefixPath(r.prefix.Path, path),
			Name: r.prefix.Name + routerName,
		},
		middlewares: r.middlewares,
		assets:      r.assets,
		routes:      r.routes,
	}
}

func (r *router) createWildcardRoute() {
	path := "/{path...}"
	for _, method := range []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete,
	} {
		r.mux.HandleFunc(
			r.createRoutePattern(method, path),
			r.createHandler(
				method, path, dynamicName, func(c Ctx) error {
					c.Response().Status(http.StatusNotFound)
					return r.core.dynamicHandler(c)
				},
			),
		)
	}
}

func (r *router) createDynamicAssetsRoute() {
	path := tempestAssetsPath + "{name}"
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		r.mux.HandleFunc(
			fmt.Sprintf("%s %s", method, path),
			r.createHandler(
				method, path, "assets", func(c Ctx) error {
					name := c.Request().PathValue("name")
					if strings.HasSuffix(name, fileSuffix.Css) && strings.Contains(name, "-"+r.assets.code+".") {
						return c.Response().Asset(name, []byte(tempest.Styles()))
					}
					if strings.HasSuffix(name, fileSuffix.Js) && strings.Contains(name, "-"+r.assets.code+".") {
						if c.Request().Header().Get("If-None-Match") == r.assets.code {
							return c.Response().Status(http.StatusNotModified).Empty()
						}
						c.Response().Header().Set(
							header.CacheControl,
							fmt.Sprintf("public, max-age=%d", int(math.Round(tempestAssetsCacheDuration.Seconds()))),
						)
						c.Response().Header().Set(
							header.ETag,
							r.assets.code,
						)
						return c.Response().Asset(name, []byte(tempest.Scripts()))
					}
					return c.Response().Status(http.StatusNotFound).Error(http.StatusText(http.StatusNotFound))
				},
			),
		)
	}
}

func (r *router) createCanonicalHandler() Handler {
	return func(c Ctx) error {
		return c.Response().Status(http.StatusPermanentRedirect).Redirect(c.Generate().Current())
	}
}

func (r *router) createCanonicalRoute(path string, lang string, config ...RouteConfig) {
	if strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	if path == "" {
		return
	}
	var name string
	methods := make([]string, 0)
	for _, cfg := range config {
		switch cfg.Type {
		case routeMethod:
			methods = cfg.Value.([]string)
		case routeName:
			name = cfg.Value.(string)
		}
	}
	if len(methods) == 0 {
		methods = append(methods, httpMethods...)
	}
	path = r.prefixPathWithLangIfEnabled(path, lang)
	if len(r.prefix.Name) > 0 {
		name = createDividedName(r.prefix.Name, name)
	}
	for _, method := range methods {
		r.mux.HandleFunc(
			fmt.Sprintf("%s %s", method, path),
			r.createHandler(method, path, name, r.createCanonicalHandler()),
		)
	}
}

func (r *router) createRoute(path string, fn Handler, lang string, config ...RouteConfig) {
	var name, layout string
	methods := make([]string, 0)
	for _, cfg := range config {
		switch cfg.Type {
		case routeMethod:
			methods = cfg.Value.([]string)
		case routeName:
			name = cfg.Value.(string)
		case routeLayout:
			layout = cfg.Value.(string)
		}
	}
	if len(layout) == 0 {
		layout = Main
	}
	if len(methods) == 0 {
		methods = append(methods, httpMethods...)
	}
	if r.prefix.Path != nil {
		switch v := r.prefix.Path.(type) {
		case string:
			path = pathx.MustJoinPath(v, path)
		case map[string]string:
			lp, ok := v[lang]
			if ok {
				path = pathx.MustJoinPath(lp, path)
			}
		}
	}
	path = r.prefixPathWithLangIfEnabled(path, lang)
	if len(r.prefix.Name) > 0 {
		name = createDividedName(r.prefix.Name, name)
	}
	matcher, pathValues := r.createMatcher(path)
	*r.routes = append(
		*r.routes, &Route{
			Lang:       lang,
			Path:       path,
			Name:       name,
			Methods:    methods,
			Layout:     r.core.layout.factories[layout],
			Matcher:    matcher,
			PathValues: pathValues,
			Firewall:   r.createFirewall(path, name),
		},
	)
	for _, method := range methods {
		r.mux.HandleFunc(
			r.createRoutePattern(method, path),
			r.createHandler(method, path, name, fn),
		)
	}
}

func (r *router) createFirewall(path, name string) []firewall.Firewall {
	result := make([]firewall.Firewall, 0)
	for _, f := range r.config.Security.Firewall {
		if f.Match(path) {
			result = append(result, f)
			continue
		}
		if f.MatchPath(path) {
			result = append(result, f)
			continue
		}
		if f.MatchGroup(name) {
			result = append(result, f)
			continue
		}
	}
	return result
}

func (r *router) createRoutePattern(method, path string) string {
	return method + " " + r.formatPatternPath(path)
}

func (r *router) createMatcher(path string) (*regexp.Regexp, []string) {
	parts := strings.Split(path, "/")
	res := make([]string, len(parts))
	pathValues := make([]string, 0)
	for i, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			res[i] = paramRegex
			pathValues = append(pathValues, part[1:len(part)-1])
			continue
		}
		res[i] = part
	}
	p := strings.Join(res, "/")
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return regexp.MustCompile("^" + p + "$"), pathValues
}

func (r *router) formatPatternPath(path string) string {
	if strings.Contains(path, "...") {
		return path
	}
	if !strings.HasSuffix(path, "/") {
		return path + "/{$}"
	}
	return path + "{$}"
}

func (r *router) mergePrefixPath(prefixPath any, path any) any {
	switch pp := prefixPath.(type) {
	case string:
		switch p := path.(type) {
		case string:
			return pp + p
		}
	case map[string]string:
		switch p := path.(type) {
		case map[string]string:
			for l, item := range pp {
				p[l] = item + p[l]
			}
			return p
		}
	}
	return path
}

func (r *router) prefixPathWithLangIfEnabled(path, lang string) string {
	if r.config.Localization.Path && !strings.HasSuffix(path, "/"+lang+"/") {
		return pathx.MustJoinPath("/"+lang+"/", path)
	}
	return path
}

func (r *router) createHandler(method, path, name string, fn Handler) func(http.ResponseWriter, *http.Request) {
	return handler{
		core:   r.core,
		method: method,
		path:   path,
		name:   name,
	}.create(fn)
}
