package main

import (
	"github.com/aellwein/slf4go"
	_ "github.com/aellwein/slf4go-native-adaptor"
	"github.com/ccremer/clustercode-worker/api"
	"github.com/ccremer/clustercode-worker/compute"
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/ccremer/clustercode-worker/process"
	"github.com/ccremer/clustercode-worker/shovel"
	"github.com/ccremer/clustercode-worker/util"
	"github.com/micro/go-config"
)

var log slf4go.Logger

func main() {
	log = slf4go.GetLogger("main")

	log.Infof("Loading configuration...")
	util.LoadConfig()

	log.SetLevel(util.StringToLogLevel(config.Get("log", "level").String("info")))
	api.Init()
	compute.Init()
	shovel.Init()
	messaging.Init()
	process.Init()
	messaging.Connect()

	computeRole := "compute"
	shovelRole := "shovel"
	role := config.Get("role").String("compute")
	if role == computeRole {
		log.Infof("Enable role: %s", computeRole)
		compute.Start()
	} else if role == shovelRole {
		log.Infof("Enable role: %s", shovelRole)
		shovel.Start()
	} else {
		util.PanicWithMessage("You need to specify this worker's role using CC_ROLE to one of [%s, %s]",
			computeRole, shovelRole)
	}

	api.StartServer()

	forever := make(chan bool)
	log.Infof("Startup complete")
	<-forever
}

