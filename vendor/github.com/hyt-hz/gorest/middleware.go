package gorest

import (
	"golang.org/x/net/context"
	"log"
	"net/http"
	"time"
)

type Middleware func(ContextHandlerFunc) ContextHandlerFunc

func LoggerHttpMiddleware(handlerFunc ContextHandlerFunc) ContextHandlerFunc {
	f := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handlerFunc(ctx, w, r)
		latency := time.Now().Sub(start)
		if len(r.URL.RawQuery) == 0 {
			log.Printf("%s %s <%.3f>", r.Method, r.URL.Path, latency.Seconds())
		} else {
			log.Printf("%s %s?%s <%.3f>", r.Method, r.URL.Path, r.URL.RawQuery, latency.Seconds())
		}
	}

	return ContextHandlerFunc(f)
}

// TODO WARNING: dangerous, should never allow all origin in production
func CORSMiddleware(handlerFunc ContextHandlerFunc) ContextHandlerFunc {
	f := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		o := r.Header.Get("Origin")
		if o == "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", o)
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept,X-Requested-With")
		handlerFunc(ctx, w, r)
	}

	return ContextHandlerFunc(f)
}

func CORSAllowCredentialsMiddleware(handlerFunc ContextHandlerFunc) ContextHandlerFunc {
	f := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		handlerFunc(ctx, w, r)
	}

	return ContextHandlerFunc(f)
}
