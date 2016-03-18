package gorest

import (
	"golang.org/x/net/context"
	"log"
	"net/http"
)

/*
way to define HTTP routing, middleware etc.

to create new routing:

r := NewRouter()
r.Get("/ualist", c.getUAListHandler)
g.Post("/ualist/ua", c.updateUAHandler)

you can define routing groups, and then define routing rule under group

g := r.NewGroup("/g/g1")
g.Get("/conference", getConferenceListHandler)

middleware can be added to routine group
g.use(LoggerHttpMiddleware)

you can chain middleware like
g.use(LoggerHttpMiddleware).use(RecoveryHttpMiddleware)

define middleware on routing rule is not possible yet


*/

type ContextHandlerFunc func(context.Context, http.ResponseWriter, *http.Request)
type ContextHandler interface {
	ServeHTTP(context.Context, http.ResponseWriter, *http.Request)
}

func HandlerFuncAdapter(hf http.HandlerFunc) ContextHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		hf(w, r)
	}
}

func HandlerAdapter(h http.Handler) ContextHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

type (
	Router struct {
		Group
		groups       []*Group
		builtEntries []routeEntry
	}

	routeEntry struct {
		method      string
		path        string
		handlerFunc ContextHandlerFunc
		handler     ContextHandler
	}

	Group struct {
		r               *Router
		parent          *Group
		child           []*Group
		path            string
		middlewareStack []Middleware
	}
)

func NewRouter() *Router {
	r := &Router{
		Group: Group{
			parent:          nil,
			child:           make([]*Group, 0, 10),
			path:            "",
			middlewareStack: make([]Middleware, 0, 10),
		},
		groups:       make([]*Group, 0, 10),
		builtEntries: make([]routeEntry, 0, 100),
	}
	r.Group.r = r
	r.groups = append(r.groups, &r.Group)

	return r
}

func (parent *Group) NewGroup(path string) *Group {

	if path == "" {
		log.Printf("group path must not be empty")
		panic("group path must not be empty")
	}

	if path[0] != '/' {
		log.Printf("group path %s not starts with '/'", path)
		panic("group path not starts with '/'")
	}

	if path == "/" {
		log.Printf("group path '/' not valid")
		panic("group path '/' not valid")
	}

	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	for _, g := range parent.child {
		if g.path == path {
			log.Printf("Duplicated group %s under %s", path, parent.path)
			panic("Duplicated group")
		}
	}

	g := &Group{
		r:               parent.r,
		parent:          parent,
		child:           make([]*Group, 0, 10),
		path:            path,
		middlewareStack: make([]Middleware, 0, 10),
	}
	g.r.groups = append(g.r.groups, g)
	parent.child = append(parent.child, g)

	return g
}

func (g *Group) Use(m Middleware) *Group {
	g.middlewareStack = append(g.middlewareStack, m)
	return g
}

func (g *Group) Get(path string, handler ContextHandlerFunc) *Group {
	g.appendEntry("GET", path, handler)
	return g
}

func (g *Group) Put(path string, handler ContextHandlerFunc) *Group {
	g.appendEntry("PUT", path, handler)
	return g
}

func (g *Group) Post(path string, handler ContextHandlerFunc) *Group {
	g.appendEntry("POST", path, handler)
	return g
}

func (g *Group) Delete(path string, handler ContextHandlerFunc) *Group {
	g.appendEntry("DELETE", path, handler)
	return g
}

func (g *Group) appendEntry(method string, path string, handlerFunc ContextHandlerFunc) {

	if handlerFunc == nil {
		log.Printf("nil handler functin for %s path %s", method, path)
		panic("nil handler functin")
	}

	if len(path) > 0 && path[0] != '/' {
		log.Printf("Invalid path %s", path)
		panic("Invalid path %s")
	}

	gr := g
	for gr != nil {
		if len(gr.path) != 0 && gr.path[len(gr.path)-1] == '/' {
			path = gr.path[:len(gr.path)-1] + path
		} else {
			path = gr.path + path
		}
		gr = gr.parent
	}
	if path == "" {
		path = "/"
	}

	corsDone := false
	for _, entry := range g.r.builtEntries {
		if path == entry.path {
			if method == entry.method {
				log.Printf("Duplicate router entry %s URL path %s", method, path)
				panic("Duplicate router entry")
			}
			if entry.method == "OPTIONS" {
				corsDone = true
				coh := entry.handler.(*corsOptionsHanlder)
				coh.addAllowedMethods(method)
			}
		}
	}

	if corsDone == false {
		coh := newCorsOptionsHandler(method)
		g.r.builtEntries = append(g.r.builtEntries, routeEntry{
			method:  "OPTIONS",
			path:    path,
			handler: coh,
		})
	}

	h := g.buildHandlerWithMiddleware(handlerFunc)

	g.r.builtEntries = append(g.r.builtEntries, routeEntry{
		method:      method,
		path:        path,
		handlerFunc: h,
	})
}

func (g *Group) buildHandlerWithMiddleware(handler ContextHandlerFunc) ContextHandlerFunc {

	h := handler

	for i := len(g.middlewareStack); i > 0; i-- {
		h = g.middlewareStack[i-1](h)
	}

	if g.parent != nil {
		return g.parent.buildHandlerWithMiddleware(h)
	} else {
		return h
	}
}
