package util

import (
    "github.com/aellwein/slf4go"
    "github.com/micro/go-config"
)

var log = slf4go.GetLogger("util")

func init() {
    log.SetLevel(StringToLogLevel(config.Get("log", "level").String("info")))
}
