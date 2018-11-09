package main

import (
	"github.com/aellwein/slf4go"
	_ "github.com/aellwein/slf4go-native-adaptor"
	"github.com/ccremer/clustercode-worker/compute"
	"github.com/ccremer/clustercode-worker/health"
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/ccremer/clustercode-worker/util"
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
)

var log slf4go.Logger

func main() {
	log = slf4go.GetLogger("main")

	log.Infof("Loading configuration...")
	config.Load(
		file.NewSource(file.WithPath("defaults.yaml")),
		file.NewSource(file.WithPath("config.yaml")),
		env.NewSource(env.WithStrippedPrefix("CC")),
	)

	log.SetLevel(util.StringToLogLevel(config.Get("log", "level").String("info")))
	health.Init()
	compute.Init()
	messaging.Init()

	computeRole := "compute"
	shovelRole := "shovel"

	messaging.Connect()

	role := config.Get("role").String("compute")
	if role == computeRole {
		compute.Start()
	} else if role == shovelRole {
		panic("Not implemented yet")
	} else {
		util.PanicWithMessage("You need to specify this worker's role using CC_ROLE to one of %s",
			computeRole, shovelRole)
	}

	health.StartServer()

	forever := make(chan bool)
	log.Infof("Startup complete")
	<-forever
}
