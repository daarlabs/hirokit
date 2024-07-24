package devtool

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"
	
	"github.com/daarlabs/hirokit/alpine"
	"github.com/daarlabs/hirokit/cookie"
	"github.com/daarlabs/hirokit/gox"
	"github.com/daarlabs/hirokit/socketer"
	"github.com/daarlabs/hirokit/tempest"
)

func HandleConnection(ws socketer.Ws) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookies := cookie.New(r, w, "/")
		id := cookies.Get(StateCookieKey)
		if len(id) == 0 {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		if err := ws.Upgrade(r, w, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func HandleRefresh(ws socketer.Ws) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws.Broadcast(
			[]byte(
				gox.Render(
					gox.Div(
						gox.Id(RefreshId),
						alpine.Init("window.location.href = window.location.href"),
					),
				),
			),
		)
	}
}

func HandleRequest(ws socketer.Ws, cache *sync.Map) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		renderTimeString := r.URL.Query().Get("renderTime")
		statusCodeString := r.URL.Query().Get("statusCode")
		statusCode, err := strconv.Atoi(statusCodeString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		renderTime, err := strconv.Atoi(renderTimeString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		props := Props{
			Path:       r.URL.Query().Get("path"),
			Name:       r.URL.Query().Get("name"),
			RenderTime: renderTime,
			StatusCode: statusCode,
			Plugin:     make(map[string][]string),
		}
		for key, plugin := range ToolConfig.Plugin {
			if v, ok := r.URL.Query()[key]; ok && len(v) > 0 {
				if _, pluginExists := props.Plugin[key]; !pluginExists {
					props.Plugin[key] = make([]string, 0)
				}
				if !plugin.Reference {
					props.Plugin[key] = v
				}
				if plugin.Reference {
					rows := make([]string, 0)
					for _, referenceKey := range v {
						if rv, rok := r.URL.Query()[referenceKey]; rok && len(rv) > 0 {
							for i, item := range rv {
								rv[i] = referenceKey + "/" + item
							}
							rows = append(rows, rv...)
						}
					}
					props.Plugin[key] = rows
				}
			}
		}
		cache.Store(id, props)
		ws.Broadcast(
			[]byte(
				gox.Render(
					createDevtoolContent(props),
				),
			),
		)
	}
}

func HandleTool(cache *sync.Map, assetsId string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookies := cookie.New(r, w, "/")
		id := cookies.Get(StateCookieKey)
		var props Props
		cachedProps, ok := cache.Load(id)
		if ok {
			props = cachedProps.(Props)
		}
		if _, err := w.Write(
			[]byte(gox.Render(
				createTool(assetsId, props),
			)),
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func HandleToolStyles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		if _, err := w.Write([]byte(tempest.Styles())); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func HandleToolScripts(assetsId string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == assetsId {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(math.Round((7*24*time.Hour).Seconds()))))
		w.Header().Set("ETag", assetsId)
		w.Header().Set("Content-Type", "application/javascript")
		if _, err := w.Write([]byte(tempest.Scripts())); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
