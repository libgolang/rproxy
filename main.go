package main

import (
	"github/libgolang/rproxy/lib"

	"github.com/libgolang/log/v2"
	"github.com/libgolang/props"
)

func main() {
	if props.IsSet("d") {
		log.SetLevel(log.LevelTrace)
	}
	configFile := props.GetProp("config")

	cfg, err := lib.ReadConfig(configFile)
	if err != nil {
		log.Fatal("%s", err)
	}

	rproxy := lib.NewReverseProxy(cfg)
	err = rproxy.ListenAndServe()
	if err != nil {
		log.Fatal("%s", err)
	}
}
