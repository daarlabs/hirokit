package hiro

import (
	"fmt"
	"net/http"
	"strings"
	
	"github.com/daarlabs/hirokit/constant/contentType"
	"github.com/daarlabs/hirokit/constant/dataType"
	"github.com/daarlabs/hirokit/constant/header"
	"github.com/daarlabs/hirokit/devtool"
	"github.com/daarlabs/hirokit/env"
	"github.com/daarlabs/hirokit/gox"
	"github.com/daarlabs/hirokit/hx"
)

type Handler func(c Ctx) error

type handler struct {
	core   *core
	method string
	path   string
	name   string
}

func (h handler) create(fn Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var writer http.ResponseWriter
		isTempest := strings.HasPrefix(r.URL.Path, tempestAssetsPath)
		compressedWriter := createCompressedWriter(w)
		matchedRoute := h.matchRoute(r)
		isWs := matchedRoute != nil && matchedRoute.Ws != nil
		if !isWs {
			writer = compressedWriter
		}
		if isWs {
			writer = w
		}
		c := createContext(
			ctxParam{
				assets:       h.core.assets,
				config:       h.core.router.config,
				layout:       h.core.layout,
				r:            r,
				w:            writer,
				matchedRoute: matchedRoute,
				routes:       h.core.router.routes,
			},
		)
		if !isWs {
			defer func(c *ctx) {
				if err := compressedWriter.Writer.Close(); err != nil {
					c.err = err
				}
			}(c)
		}
		if h.core.router.config.Router.Recover {
			defer h.createRecover(c)
		}
		if !isTempest {
			for _, middleware := range h.applyInternalMiddlewares(matchedRoute, h.core.router.middlewares) {
				c.err = middleware(c)
				if c.err != nil && !isWs {
					h.createResponse(c)
					return
				}
				if c.err != nil && isWs {
					panic(c.err)
				}
			}
		}
		if len(c.response.DataType) == 0 && fn != nil {
			err := fn(c)
			if err != nil {
				c.err = err
			}
		}
		if isWs {
			h.upgradeResponse(c, writer, r, matchedRoute)
			return
		}
		if env.Development() && !isTempest && h.core.config.Dev.Tool {
			devtoolPush(c)
		}
		h.createResponse(c)
	}
}

func (h handler) applyInternalMiddlewares(matchedRoute *Route, middlewares []Handler) []Handler {
	r := make([]Handler, 0)
	r = append(r, createFormMiddleware())
	if h.core.config.Localization.Enabled {
		r = append(r, createLangMiddleware())
	}
	if h.core.config.Security.Csrf != nil {
		r = append(r, createCsrfMiddleware())
	}
	if matchedRoute != nil && len(matchedRoute.Firewall) > 0 {
		r = append(r, createFirewallMiddleware(matchedRoute.Firewall))
	}
	r = append(r, middlewares...)
	return r
}

func (h handler) upgradeResponse(c *ctx, w http.ResponseWriter, r *http.Request, matchedRoute *Route) {
	var clientId string
	sessionExists := c.Auth().Session().MustExists()
	if !sessionExists {
		clientId = fmt.Sprint(c.Auth().Session().MustGet().Id)
	}
	if sessionExists {
		clientId = fmt.Sprint(c.Auth().Session().MustGet().Id)
	}
	matchedRoute.Ws.MustUpgrade(r, w, clientId)
}

func (h handler) createResponse(c *ctx) {
	if c.response.DataType != dataType.Asset {
		c.w.Header().Set(header.CacheControl, "no-cache, no-store, must-revalidate")
	}
	if c.err != nil {
		c.w.Header().Set(header.ContentType, contentType.Text)
		if c.response.StatusCode == http.StatusOK {
			c.w.WriteHeader(http.StatusInternalServerError)
		}
		if c.response.StatusCode != http.StatusOK {
			c.w.WriteHeader(c.response.StatusCode)
		}
		_, err := c.w.Write([]byte(c.err.Error()))
		if err != nil {
			http.Error(c.w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		return
	}
	if c.response.DataType == dataType.Redirect {
		if c.Request().Is().Hx() {
			c.w.Header().Set(hx.ResponseHeaderRedirect, c.response.Value)
			c.w.WriteHeader(http.StatusOK)
			return
		}
		if c.response.StatusCode == http.StatusOK {
			c.response.StatusCode = http.StatusFound
		}
		http.Redirect(c.w, c.r, c.response.Value, c.response.StatusCode)
		return
	}
	c.w.Header().Set(header.ContentLength, "")
	if c.response.DataType == dataType.Stream {
		c.w.Header().Set(header.ContentDisposition, fmt.Sprintf("attachment;filename=%s", c.response.Value))
	}
	if c.response.DataType == dataType.Asset {
		c.w.Header().Set(header.AcceptRanges, "bytes")
	}
	if c.response.DataType != dataType.Empty {
		c.w.Header().Set(header.ContentType, c.response.ContentType)
	}
	c.w.WriteHeader(c.response.StatusCode)
	if c.response.DataType == dataType.Empty {
		return
	}
	if _, c.err = c.w.Write(c.response.Bytes); c.err != nil {
		http.Error(c.w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h handler) createRecover(c *ctx) {
	if e := recover(); e != nil {
		err, ok := e.(error)
		if !ok {
			http.Error(c.w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if c.response.StatusCode == http.StatusOK || c.response.StatusCode == http.StatusBadRequest {
			c.response.StatusCode = http.StatusInternalServerError
		}
		c.err = err
		isHx := c.Request().Is().Hx()
		if env.Development() && !isHx {
			err = c.Response().Html(gox.Render(devtool.Recover(c.Generate().Assets(c.Request().Name()), err)))
		}
		if !env.Development() || isHx {
			err = h.core.dynamicHandler(c)
		}
		c.err = nil
		if err != nil {
			c.err = err
		}
		h.createResponse(c)
	}
}

func (h handler) matchRoute(r *http.Request) *Route {
	path := r.URL.Path
	if strings.HasPrefix(path, tempestAssetsPath) {
		return nil
	}
	matches := make([]*Route, 0)
	for _, match := range *h.core.router.routes {
		if match.Matcher.MatchString(path) {
			matches = append(matches, match)
		}
	}
	for _, match := range matches {
		var matchPath string
		for _, pathValue := range match.PathValues {
			matchPath = strings.Replace(match.Path, "{"+pathValue+"}", r.PathValue(pathValue), 1)
		}
		if strings.TrimSuffix(matchPath, "/") == strings.TrimSuffix(path, "/") {
			return match
		}
	}
	if len(matches) > 0 {
		return matches[0]
	}
	return nil
}
