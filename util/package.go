package util

import (
	"github.com/aellwein/slf4go"
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
)

var log = slf4go.GetLogger("util")

func init() {
	log.SetLevel(StringToLogLevel(config.Get("log", "level").String("info")))
}

func LoadConfig() {
	config.Load(
		file.NewSource(file.WithPath("defaults.yaml")),
		file.NewSource(file.WithPath("config.yaml")),
		env.NewSource(env.WithStrippedPrefix("CC")),
	)
}
