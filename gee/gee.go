package gee

import (
	"log"
	"net/http"
)

type Engine struct {
	// router承担路由功能
	router *router
	groups []*RouterGroup
	*RouterGroup
}

type RouterGroup struct {
	engine      *Engine
	parent      *RouterGroup
	middlewares []HandlerFunc
	// eg prefix /api /host
	prefix string
}

var _ http.Handler = &Engine{}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = append(engine.groups, engine.RouterGroup)
	return engine
}

func (e *Engine) addRoute(method string, path string, handler HandlerFunc) {
	e.router.addRoute(method, path, handler)
}

func (e *Engine) GET(path string, handler HandlerFunc) {
	e.addRoute("GET", path, handler)
}

func (e *Engine) POST(path string, handler HandlerFunc) {
	e.addRoute("POST", path, handler)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(w, r)
	e.router.handle(ctx)
}

func (e *Engine) Run(port string) {
	log.Fatal(http.ListenAndServe(port, e))
}

func (rg *RouterGroup) Group(prefix string) *RouterGroup {
	engine := rg.engine
	newRG := &RouterGroup{
		engine: rg.engine,
		parent: rg,
		prefix: rg.prefix + prefix,
	}
	engine.groups = append(engine.groups, newRG)
	return newRG
}

func (rg *RouterGroup) addRoute(method string, part string, handler HandlerFunc) {
	rg.engine.addRoute(method, rg.prefix+part, handler)
}

func (rg *RouterGroup) GET(part string, handler HandlerFunc) {
	rg.addRoute("GET", part, handler)
}

func (rg *RouterGroup) POST(part string, handler HandlerFunc) {
	rg.addRoute("POST", part, handler)
}
