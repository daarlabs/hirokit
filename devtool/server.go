package devtool

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sync"
	
	"github.com/dchest/uniuri"
	
	"github.com/daarlabs/hirokit/socketer"
)

//go:embed static
var static embed.FS

func Serve() {
	fsys, err := fs.Sub(static, "static")
	if err != nil {
		log.Fatalln(err)
	}
	assetsId := uniuri.New()
	cache := new(sync.Map)
	ws := socketer.New()
	mux := http.NewServeMux()
	mux.Handle("GET /.dev/static/", http.StripPrefix("/.dev/static/", http.FileServer(http.FS(fsys))))
	mux.HandleFunc("GET /.dev/tool/{$}", createMiddleware(HandleTool(cache, assetsId)))
	mux.HandleFunc(fmt.Sprintf("GET /.dev/.tempest/styles.%s.css", assetsId), HandleToolStyles())
	mux.HandleFunc(fmt.Sprintf("GET /.dev/.tempest/scripts.%s.js", assetsId), HandleToolScripts(assetsId))
	mux.HandleFunc("GET /.dev/refresh/{$}", createMiddleware(HandleRefresh(ws)))
	mux.HandleFunc("GET /.dev/push/{id}/{$}", createMiddleware(HandleRequest(ws, cache)))
	mux.HandleFunc("GET /.dev/{$}", createMiddleware(HandleConnection(ws)))
	fmt.Println("dev-server is running on :", ToolConfig.Port)
	log.Fatalln(http.ListenAndServe(":"+ToolConfig.Port, mux))
}
