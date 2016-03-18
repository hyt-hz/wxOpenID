package main

import (
	"flag"
	"github.com/hyt-hz/wxOpenID/log"
	"github.com/hyt-hz/wxOpenID/utils"
)

func main() {

	options := Option{}

	configPath := flag.String("c", "conf", "Config file path, can be either file or directory")
	flag.Parse()

	//start log
	log.Start("conf/wx.log.xml")
	defer log.Flush()

	if err := utils.ParseConfigFile(*configPath, "wx.yaml", &options); err != nil {
		log.Critical("Failed to parse config file")
		return
	}
	log.Info("Config options:%+v", options)

	s, err := NewServer(&options)
	if err != nil {
		log.Critical("Failed to create server: %s", err)
	}

	s.Run()
}
