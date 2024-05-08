package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

type Engine struct {
	// router承担路由功能
	router *router
	groups []*RouterGroup
	*RouterGroup
	htmlTemplates *template.Template
	funcMap       template.FuncMap
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

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
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

// Actual request process entrance.
// Called during every request.
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range e.groups {
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	ctx := newContext(w, r)
	ctx.handlers = middlewares
	ctx.engine = e
	e.router.handle(ctx)
}

func (e *Engine) Run(port string) {
	log.Fatal(http.ListenAndServe(port, e))
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

func (e *Engine) LoadHTMLGlob(pattern string) {
	e.htmlTemplates = template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
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

func (rg *RouterGroup) Use(middlewares ...HandlerFunc) {
	rg.middlewares = append(rg.middlewares, middlewares...)
}

func (rg *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(rg.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

// serve static files
func (rg *RouterGroup) Static(relativePath string, root string) {
	handler := rg.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	rg.GET(urlPattern, handler)
}
