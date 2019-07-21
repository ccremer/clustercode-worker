package main

import (
	"fmt"
	"github.com/ccremer/clustercode-worker/api"
	"github.com/ccremer/clustercode-worker/compute"
	"github.com/ccremer/clustercode-worker/config"
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/ccremer/clustercode-worker/shovel"
	log "github.com/sirupsen/logrus"
	"strings"
)

var Version string
var Commit string

func main() {
	if err := config.LoadConfig(); err != nil {
		log.
			WithError(err).
			WithField("help", "Be sure to NOT specify the file extension.").
			Error("Could not load config.")
	}
	config.ConfigureLogging()

	cfg := config.GetConfig()
	config.SaveConfig()

	role := cfg.Role
	if role != config.RoleShovel && role != config.RoleCompute {
		log.WithFields(log.Fields{
			"variable": "role",
			"help": fmt.Sprintf("Specify the role either in CC_ROLE, by providing the CLI flag --role, "+
				"or in a config file. Allowed values: [%s]",
				strings.Join([]string{config.RoleCompute, config.RoleShovel}, ",")),
		}).Fatal("Correct role specification of this worker is required.")
	}

	service := messaging.NewRabbitMqService(cfg.RabbitMq.Url)
	api.StartHttpServer(service)
	service.Start()

	if role == config.RoleCompute {
		compute.NewComputeInstance(service)
	} else if role == config.RoleShovel {
		shovel.NewInstance(service)
	}

	forever := make(chan bool)
	log.WithFields(log.Fields{
		"version": Version,
		"commit":  Commit,
		"role":    role,
	}).Info("Startup complete.")
	<-forever
}
