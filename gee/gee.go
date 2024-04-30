package gee

import (
	"log"
	"net/http"
)

type Engine struct {
	// router承担路由功能
	router *router
}

var _ http.Handler = &Engine{}

func New() *Engine {
	return &Engine{router: newRouter()}
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
