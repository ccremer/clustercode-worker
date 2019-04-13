package main

import (
	"fmt"
	"github.com/ccremer/clustercode-worker/api"
	"github.com/ccremer/clustercode-worker/compute"
	"github.com/ccremer/clustercode-worker/config"
	"github.com/ccremer/clustercode-worker/entities"
	"github.com/ccremer/clustercode-worker/messaging"
	"github.com/ccremer/clustercode-worker/schema"
	"github.com/ccremer/clustercode-worker/shovel"
	log "github.com/sirupsen/logrus"
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

	service := messaging.NewRabbitMqService(cfg.RabbitMq.Url)
	service.Start()
	entities.Validator = schema.NewXmlValidator(cfg.Api.Schema.Path)

	computeRole := "compute"
	shovelRole := "shovel"
	role := cfg.Role
	if role == computeRole {
		compute.NewComputeInstance(service)
	} else if role == shovelRole {
		shovel.NewInstance(service)
	} else {
		log.WithFields(log.Fields{
			"variable": "role",
			"help": fmt.Sprintf("Specify the role either in CC_ROLE, by providing the CLI flag --role, "+
				"or in a config file. Allowed values: %s", []string{computeRole, shovelRole}),
		}).Fatal("Correct role specification of this worker is required.")
	}

	api.StartHttpServer(service)

	forever := make(chan bool)
	log.WithFields(log.Fields{
		"version": Version,
		"commit": Commit,
		"role":    role,
	}).Info("Startup complete.")
	<-forever
}
