package gorest

import (
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

type corsOptionsHanlder struct {
	allowedMethods     []string
	allowMethodsHeader string
}

func newCorsOptionsHandler(allowedMethods ...string) *corsOptionsHanlder {
	coh := &corsOptionsHanlder{
		allowedMethods: make([]string, 0, 4),
	}
	coh.addAllowedMethods(allowedMethods...)
	return coh
}

func (coh *corsOptionsHanlder) addAllowedMethods(allowedMethods ...string) {
	for _, method := range allowedMethods {
		coh.allowedMethods = append(coh.allowedMethods, strings.ToUpper(method))
		coh.allowMethodsHeader = strings.Join(coh.allowedMethods, ", ")
	}
}

func (coh *corsOptionsHanlder) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	testMethod := r.Header.Get("Access-Control-Request-Method")
	for _, allowedMethod := range coh.allowedMethods {
		if allowedMethod == testMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", coh.allowMethodsHeader)
			w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			return
		}
	}

	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	return
}
