package compute

import (
	"github.com/aellwein/slf4go"
	"github.com/ccremer/clustercode-worker/util"
	"github.com/micro/go-config"
)

var log = slf4go.GetLogger("compute")

func Init() {
	log.SetLevel(util.StringToLogLevel(config.Get("log", "level").String("info")))
}
