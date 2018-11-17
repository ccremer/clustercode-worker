package main

import (
    "github.com/aellwein/slf4go"
    _ "github.com/aellwein/slf4go-native-adaptor"
    "github.com/ccremer/clustercode-worker/api"
    "github.com/ccremer/clustercode-worker/compute"
    "github.com/ccremer/clustercode-worker/messaging"
    "github.com/ccremer/clustercode-worker/shovel"
    "github.com/ccremer/clustercode-worker/util"
    "github.com/micro/go-config"
    "github.com/micro/go-config/source/env"
    "github.com/micro/go-config/source/file"
)

var log slf4go.Logger

func main() {
	log = slf4go.GetLogger("main")

	LoadConfig()

	log.SetLevel(util.StringToLogLevel(config.Get("log", "level").String("info")))
	api.Init()
	compute.Init()
	shovel.Init()
	messaging.Init()
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

func LoadConfig() {
    log.Infof("Loading configuration...")
    config.Load(
        file.NewSource(file.WithPath("defaults.yaml")),
        file.NewSource(file.WithPath("config.yaml")),
        env.NewSource(env.WithStrippedPrefix("CC")),
    )

}
