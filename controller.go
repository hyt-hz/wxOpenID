package main

import (
	"github.com/hyt-hz/gorest"
	"github.com/hyt-hz/wxOpenID/service"
	"golang.org/x/net/context"
	"net/http"
)

type controller struct {
	wxs *service.WXService
}

func NewController(hbgroup *gorest.Group, wxs *service.WXService) (c *controller, err error) {

	c = &controller{
		wxs: wxs,
	}

	// register HTTP middlewares and handlers
	//g.Use(cm.auth.AuthUserHttpMiddleware)

	hbgroup.Get("/validateServer", c.validateServer)

	return
}

func (c *controller) validateServer(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	signature := r.FormValue("signature")
	timestamp := r.FormValue("timestamp")
	nonce := r.FormValue("nonce")
	echostr := r.FormValue("echostr")

	if signature == "" || timestamp == "" || nonce == "" || echostr == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	c.wxs.ValidateServer(nil)
	w.Write([]byte(echostr))
	return
}
