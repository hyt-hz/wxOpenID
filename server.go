package main

import (
	"github.com/hyt-hz/gorest"
	"github.com/hyt-hz/wxOpenID/log"
	"github.com/hyt-hz/wxOpenID/service"
	"net/http"
)

type Option struct {
	Listen string
}

type server struct {
	option Option
}

var APIPrefix = "/wx"

func NewServer(option *Option) (s *server, err error) {

	s = &server{
		option: *option,
	}

	r := gorest.NewRouter()
	r.Use(gorest.LoggerHttpMiddleware)
	r.Use(gorest.RecoveryHttpMiddleware)
	r.Use(gorest.CORSMiddleware)

	hbs := service.NewWXService()

	// hongbao manager related API
	hongbaoGroup := r.NewGroup(APIPrefix)
	NewController(hongbaoGroup, hbs)

	router := gorest.BindHttprouter(r)

	http.Handle("/", router)

	return
}

func (s *server) Run() error {

	log.Info("Start to listen on %s", s.option.Listen)
	err := http.ListenAndServe(s.option.Listen, nil)
	if err != nil {
		log.Error("ListenAndServe: %v", err)
	}

	return err
}
