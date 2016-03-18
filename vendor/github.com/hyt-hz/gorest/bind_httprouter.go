package gorest

import (
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
	"net/http"
)

type paramsCtxKey int

var key = paramsCtxKey(0)

func GetParams(ctx context.Context) (p httprouter.Params, ok bool) {
	p, ok = ctx.Value(key).(httprouter.Params)
	return
}

func wrapperHandlerFunc(h ContextHandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, key, p)
		h(ctx, w, r)
	}
}

func wrapperHandler(h ContextHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, key, p)
		h.ServeHTTP(ctx, w, r)
	}
}

func BindHttprouter(r *Router) *httprouter.Router {
	hr := httprouter.New()

	for _, e := range r.builtEntries {
		if e.handlerFunc != nil {
			hr.Handle(e.method, e.path, wrapperHandlerFunc(e.handlerFunc))
		} else {
			hr.Handle(e.method, e.path, wrapperHandler(e.handler))
		}
	}

	return hr
}
