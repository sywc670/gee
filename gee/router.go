package gee

import (
	"net/http"
	"strings"
)

type HandlerFunc func(*Context)

type router struct {
	// handlers key eg, handlers['GET-/p/:lang/doc'], handlers['POST-/p/book']
	handlers map[string]HandlerFunc
	// roots key eg, roots['GET'] roots['POST']
	roots map[string]*node
}

func newRouter() *router {
	return &router{handlers: make(map[string]HandlerFunc), roots: make(map[string]*node)}
}

func parsePattern(pattern string) (parts []string) {
	patternSlice := strings.Split(pattern, "/")
	for _, part := range patternSlice {
		if part != "" {
			parts = append(parts, part)
			if part[0] == '*' {
				break
			}
		}
	}
	return
}

// register route and handler
func (r *router) addRoute(method string, path string, handler HandlerFunc) {
	parts := parsePattern(path)
	key := method + "-" + path

	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}

	r.roots[method].insert(path, parts, 0)
	r.handlers[key] = handler
}

// match route
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}
	parts := parsePattern(path)
	n := root.search(parts, 0)

	if n != nil {
		nodeParts := parsePattern(n.pattern)
		params := make(map[string]string)
		for index, nodePart := range nodeParts {
			if nodePart[0] == ':' {
				params[nodePart[1:]] = parts[index]
			}
			if nodePart[0] == '*' && len(nodePart) > 1 {
				params[nodePart[1:]] = strings.Join(parts[index:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		// path是实际路径，pattern是匹配路径 eg path:/lang/go pattern:/lang/:name
		key := c.Method + "-" + n.pattern
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next()
}
