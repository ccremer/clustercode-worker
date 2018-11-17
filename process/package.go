package process

import (
	"github.com/aellwein/slf4go"
	_ "github.com/aellwein/slf4go-native-adaptor"
	"github.com/ccremer/clustercode-worker/util"
	"github.com/micro/go-config"
)

var log = slf4go.GetLogger("process")

func Init() {
	log.SetLevel(util.StringToLogLevel(config.Get("log", "level").String("info")))
}
